package controllers

import (
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

var _ = Describe("Power outlet controller", func() {
	Context("Power outlet controller resource test", func() {

		testName := "test-outlet-controller-test" + rand.String(4)

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

		It("should create and delete power outlet resource in namespace (KubeAPI test)", func() {

			By("Non existent resource 'Poweroutlet' is expected")
			// TODO lecture ---> type of object var has to be pointer to get updates from k8sclient, normal struct will fail

			// TODO lecture ---> remove Switch to see defaulting
			powerOutlet := &personaliotv1alpha1.Poweroutlet{
				Spec: personaliotv1alpha1.PoweroutletSpec{
					Switch:           internal.PowerOffSignal,
					OutletName:       "light-one",
					MQTTStatusTopik:  "stat/gosund_p1_1_12FCA5/POWER1",
					MQTTCommandTopik: "cmnd/gosund_p1_1_12FCA5/POWER1",
				},
			}

			err := k8sClient.Get(ctx, typedNs, powerOutlet)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			powerOutletList := &personaliotv1alpha1.PoweroutletList{}
			//err = k8sClient.List(ctx, powerOutletList, &client.ListOptions{Namespace: testName})
			Expect(k8sClient.List(ctx, powerOutletList, &client.ListOptions{Namespace: testName})).Should(Succeed())
			Expect(len(powerOutletList.Items)).To(BeIdenticalTo(0))

			By("Create a power outlet resource")
			powerOutlet.Namespace = testName
			// TODO talk:
			// first: leave outlet name empty
			// second: set outlet to wrong name powerOutlet.Name = "Test-Power-Outlet"
			powerOutlet.Name = "light-one"
			err = k8sClient.Create(ctx, powerOutlet)
			Expect(err).ToNot(HaveOccurred())

			// TODO Questions to explain: How is the test local kubeapi reached, how is the test local etcd inspected?

			By("Power outlet object should be found.")

			Eventually(func() error {
				return k8sClient.List(ctx, powerOutletList, &client.ListOptions{Namespace: testName})
			}, time.Minute, time.Second).Should(Succeed())
			Expect(len(powerOutletList.Items)).To(BeIdenticalTo(1))

			powerOutletKey := client.ObjectKeyFromObject(powerOutlet)
			err = k8sClient.Get(ctx, powerOutletKey, powerOutlet)
			Expect(err).ShouldNot(HaveOccurred())

			// TODO talk:
			// Next test will fail w/o defaulting

			By("A newly created power outlet switch status is not set.")
			Expect(powerOutlet.Status.Switch).To(BeIdenticalTo(""))

			By("Reconciling is expected to run w/o error and the status field switch is set.")

			var publisher mqttiot.MQTTPublisher
			var subscriber mqttiot.MQTTSubscriber

			publisher = &mqttiot.FakeMQTTPublisher{
				ConnectError: nil,
				PublishError: nil,
			}
			subscriber = &mqttiot.FakeMQTTSubscriber{
				ConnectError:     nil,
				SubscribeError:   nil,
				UnsubscribeError: nil,
				ExpectedMessages: []mqttiot.MQTTMessage{{
					Topik:     powerOutlet.Spec.MQTTStatusTopik,
					Msg:       internal.PowerOffSignal,
					Duplicate: false,
				}},
			}

			powerOutletController := &PoweroutletReconciler{
				Client:         k8sClient,
				Scheme:         k8sClient.Scheme(),
				MQTTPublisher:  publisher,
				MQTTSubscriber: subscriber,
			}

			_, err = powerOutletController.Reconcile(ctx, reconcile.Request{
				NamespacedName: powerOutletKey,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("reconciled power outlet should have status switch set to 'off'")

			Eventually(func() error {
				return k8sClient.Get(ctx, powerOutletKey, powerOutlet)
			}, time.Minute, time.Second).Should(Succeed())
			Expect(powerOutlet.Status.Switch).To(BeIdenticalTo(powerOutlet.Spec.Switch))

			By("Updated power outlet should be updated processed by reconciler")
			powerOutlet.Spec.Switch = internal.PowerOnSignal
			err = k8sClient.Update(ctx, powerOutlet)

			Eventually(func() error {
				return k8sClient.Get(ctx, powerOutletKey, powerOutlet)
			}, time.Minute, time.Second).Should(Succeed())
			Expect(powerOutlet.Spec.Switch).To(BeIdenticalTo(internal.PowerOnSignal))
			Expect(powerOutlet.Status.Switch).To(BeIdenticalTo(internal.PowerOffSignal))

			subscriber = &mqttiot.FakeMQTTSubscriber{
				ConnectError:     nil,
				SubscribeError:   nil,
				UnsubscribeError: nil,
				ExpectedMessages: []mqttiot.MQTTMessage{{
					Topik:     powerOutlet.Spec.MQTTStatusTopik,
					Msg:       internal.PowerOffSignal,
					Duplicate: false,
				}, {
					Topik:     powerOutlet.Spec.MQTTStatusTopik,
					Msg:       internal.PowerOnSignal,
					Duplicate: false,
				}},
			}
			powerOutletController.MQTTSubscriber = subscriber

			_, err = powerOutletController.Reconcile(ctx, reconcile.Request{
				NamespacedName: powerOutletKey,
			})
			Expect(err).To(Not(HaveOccurred()))

			Eventually(func() error {
				return k8sClient.Get(ctx, powerOutletKey, powerOutlet)
			}, time.Minute, time.Second).Should(Succeed())
			Expect(powerOutlet.Status.Switch).To(BeIdenticalTo(internal.PowerOnSignal))

			By("Delete the created power outlet resource")
			err = k8sClient.Delete(ctx, powerOutlet)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(ctx, typedNs, powerOutlet)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})
})
