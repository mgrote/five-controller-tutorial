package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
	"github.com/mgrote/personal-iot/internal"
	"github.com/mgrote/personal-iot/internal/mqttiot"
)

var _ = Describe("Power strip controller", func() {
	Context("Power strip controller resource test", func() {

		testName := "test-powerstrip-controller-test" + rand.String(4)

		ctx := context.Background()

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: testName,
			},
		}

		typedNs := types.NamespacedName{Name: testName, Namespace: testName}

		BeforeEach(func() {
			By("Create Namespace to perform test")
			err := k8sClient.Create(ctx, &ns)
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("Delete Namespace to clean up test")
			_ = k8sClient.Delete(ctx, &ns)
		})

		It("should create power strip resource in namespace", func() {

			By("Non existent resource 'Powerstrip' is expected")
			powerStrip := &personaliotv1alpha1.Powerstrip{}
			err := k8sClient.Get(ctx, typedNs, powerStrip)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			powerStripList := &personaliotv1alpha1.PowerstripList{}
			Expect(k8sClient.List(ctx, powerStripList, &client.ListOptions{Namespace: testName})).Should(Succeed())
			// --------------------------------------------------------------------------------------------

			By("Create poweroutlets and powerstrip.")

			powerOutletOne := &personaliotv1alpha1.Poweroutlet{
				Spec: personaliotv1alpha1.PoweroutletSpec{
					Switch:           internal.PowerOffSignal,
					OutletName:       "light-one",
					MQTTStatusTopik:  "stat/gosund_p1_1_12FCA5/POWER1",
					MQTTCommandTopik: "cmnd/gosund_p1_1_12FCA5/POWER1",
				},
			}

			powerOutletTwo := &personaliotv1alpha1.Poweroutlet{
				Spec: personaliotv1alpha1.PoweroutletSpec{
					Switch:           internal.PowerOffSignal,
					OutletName:       "light-two",
					MQTTStatusTopik:  "stat/gosund_p1_1_12FCA5/POWER2",
					MQTTCommandTopik: "cmnd/gosund_p1_1_12FCA5/POWER2",
				},
			}

			powerOutletThree := &personaliotv1alpha1.Poweroutlet{
				Spec: personaliotv1alpha1.PoweroutletSpec{
					Switch:           internal.PowerOffSignal,
					OutletName:       "light-three",
					MQTTStatusTopik:  "stat/gosund_p1_1_12FCA5/POWER3",
					MQTTCommandTopik: "cmnd/gosund_p1_1_12FCA5/POWER3",
				},
			}
			outlets := []*personaliotv1alpha1.Poweroutlet{powerOutletOne, powerOutletTwo, powerOutletThree}

			// location name will be reused later on
			locationName := "here"
			// setup power strip
			powerStrip.Name = "light-strip"
			powerStrip.Namespace = testName
			powerStrip.Spec.Outlets = outlets
			powerStrip.Spec.LocationName = locationName
			powerStrip.Spec.MQTTStateTopik = "tele/gosund_p1_1_12FCA5/STATE"
			powerStrip.Spec.MQTTTelemetryTopik = "tele/gosund_p1_1_12FCA5/SENSOR"
			err = k8sClient.Create(ctx, powerStrip)
			Expect(err).ToNot(HaveOccurred())

			powerStripKey := client.ObjectKeyFromObject(powerStrip)
			Eventually(func() error {
				return k8sClient.Get(ctx, powerStripKey, powerStrip)
			}, time.Minute, time.Second).Should(Succeed())

			Expect(err).ShouldNot(HaveOccurred())

			var subscriber mqttiot.MQTTSubscriber
			subscriber = &mqttiot.FakeMQTTSubscriber{
				ConnectError:     nil,
				SubscribeError:   nil,
				UnsubscribeError: nil,
				ExpectedMessages: []mqttiot.MQTTMessage{{
					Topik:     "someIgnoredTopik",
					Msg:       internal.PowerOffSignal,
					Duplicate: false,
				}},
			}

			powerStripController := &PowerstripReconciler{
				Client:         k8sClient,
				Scheme:         k8sClient.Scheme(),
				MQTTSubscriber: subscriber,
			}

			_, err = powerStripController.Reconcile(ctx, reconcile.Request{
				NamespacedName: powerStripKey,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Three power outlet object should be found after reconciliation.")

			powerOutletList := &personaliotv1alpha1.PoweroutletList{}
			Eventually(func() error {
				return k8sClient.List(ctx, powerOutletList, &client.ListOptions{Namespace: testName})
			}, time.Minute, time.Second).Should(Succeed())
			Expect(len(powerOutletList.Items)).To(BeIdenticalTo(3))
			containedItemNames := [3]string{}
			for i, outlet := range powerOutletList.Items {
				containedItemNames[i] = outlet.Name
			}
			Expect(containedItemNames).To(ContainElements("light-one", "light-two", "light-three"))

			By("A location object should be found after reconciliation")

			// note the missing metav1.TypeMeta in object
			location := &personaliotv1alpha1.Location{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: testName, Name: locationName}, location)
			}, time.Minute, time.Second).Should(Succeed())
			// note the filled metav1.TypeMeta in list
			locationList := &personaliotv1alpha1.LocationList{}
			Eventually(func() error {
				return k8sClient.List(ctx, locationList, &client.ListOptions{Namespace: testName})
			}, time.Minute, time.Second).Should(Succeed())

			By("The power strip status should be set up")

			Eventually(func() error {
				return k8sClient.Get(ctx, powerStripKey, powerStrip)
			}, time.Minute, time.Second).Should(Succeed())
			Expect(powerStrip.Status.Location).To(BeIdenticalTo(locationName))
			Expect(len(powerStrip.Status.Outlets)).To(BeIdenticalTo(3))
			Expect(powerStrip.Status.Outlets).To(ContainElements("light-one", "light-two", "light-three"))
		})

	})
})
