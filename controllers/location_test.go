package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
)

var _ = Describe("Location integration", func() {
	Context("Location integration resource test", func() {

		testName := "test-location-test" + rand.String(4)

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

		It("should create location in namespace", func() {
			By("Non existent resource 'Location' is expected.")
			locationOne := personaliotv1alpha1.Location{}
			err := k8sClient.Get(ctx, typedNs, &locationOne)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			locationList := personaliotv1alpha1.LocationList{}
			Expect(k8sClient.List(ctx, &locationList, &client.ListOptions{Namespace: testName})).Should(Succeed())

		})
	})
})
