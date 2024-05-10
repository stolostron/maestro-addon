package e2e_test

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"

	workv1 "open-cluster-management.io/api/work/v1"
	workv1alpha1 "open-cluster-management.io/api/work/v1alpha1"
)

var _ = ginkgo.Describe("Loopback Test", func() {
	ginkgo.Context("ManifestWorkreplicaSet CRUD", func() {
		var mwrsName string
		var manifestNamespace string
		var manifestName string

		ginkgo.BeforeEach(func() {
			mwrsName = fmt.Sprintf("mwrs-%s", rand.String(5))
			manifestNamespace = fmt.Sprintf("mwrs-test-%s", rand.String(5))
			manifestName = "busybox"
		})

		ginkgo.It("Should be able to create/retrieve/update/delete a ManifestWorkreplicaSet successfully", func() {
			ginkgo.By("create a manifestworkreplicaset", func() {
				manifests := []workv1.Manifest{
					toManifest(newNamespace(manifestNamespace)),
					toManifest(newDeployment(manifestNamespace, manifestName, 1)),
				}
				_, err := workClient.WorkV1alpha1().ManifestWorkReplicaSets(defaultNamespace).Create(
					context.Background(), newManifestWorkReplicaSet(mwrsName, manifests), metav1.CreateOptions{})
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				gomega.Eventually(func() error {
					mwrs, err := workClient.WorkV1alpha1().ManifestWorkReplicaSets(defaultNamespace).Get(
						context.Background(), mwrsName, metav1.GetOptions{})
					if err != nil {
						return err
					}

					if !meta.IsStatusConditionTrue(
						mwrs.Status.Conditions, workv1alpha1.ManifestWorkReplicaSetConditionManifestworkApplied) {
						return fmt.Errorf("unexpected condition: %v", mwrs.Status.Conditions)
					}

					if mwrs.Status.Summary.Available != 1 {
						return fmt.Errorf("unexpected summary: %v", mwrs.Status.Conditions)
					}

					return nil
				}, timeout, time.Second).Should(gomega.Succeed())
			})

			ginkgo.By("the manifests of the manifestworkreplicaset should be created on the managed cluster", func() {
				gomega.Eventually(func() error {
					_, err := spokeKubeClient.CoreV1().Namespaces().Get(
						context.Background(), manifestNamespace, metav1.GetOptions{})
					if err != nil {
						return err
					}

					deploy, err := spokeKubeClient.AppsV1().Deployments(manifestNamespace).Get(
						context.Background(), manifestName, metav1.GetOptions{})
					if err != nil {
						return err
					}

					if *deploy.Spec.Replicas != 1 {
						return fmt.Errorf("expected replicas 1, but got %d", *deploy.Spec.Replicas)
					}

					return nil
				}, timeout, time.Second).Should(gomega.Succeed())
			})

			ginkgo.By("the works of the manifestworkreplicaset should not be created", func() {
				works, err := workClient.WorkV1().ManifestWorks(clusterName).List(
					context.Background(), metav1.ListOptions{
						LabelSelector: fmt.Sprintf("work.open-cluster-management.io/manifestworkreplicaset=%s.%s",
							defaultNamespace, mwrsName),
					})
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(len(works.Items)).To(gomega.Equal(0))
			})

			ginkgo.By("update the manifestworkreplicaset", func() {
				err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					mwrs, err := workClient.WorkV1alpha1().ManifestWorkReplicaSets(defaultNamespace).Get(
						context.Background(), mwrsName, metav1.GetOptions{})
					gomega.Expect(err).ToNot(gomega.HaveOccurred())

					newMWRS := mwrs.DeepCopy()
					newMWRS.Spec.ManifestWorkTemplate.Workload.Manifests = []workv1.Manifest{
						toManifest(newNamespace(manifestNamespace)),
						toManifest(newDeployment(manifestNamespace, manifestName, 2)),
					}

					_, err = workClient.WorkV1alpha1().ManifestWorkReplicaSets(defaultNamespace).Update(
						context.Background(), newMWRS, metav1.UpdateOptions{})
					return err
				})
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})

			ginkgo.By("the manifests of the manifestworkreplicaset should be updated on the managed cluster", func() {
				gomega.Eventually(func() error {
					deploy, err := spokeKubeClient.AppsV1().Deployments(manifestNamespace).Get(
						context.Background(), manifestName, metav1.GetOptions{})
					if err != nil {
						return err
					}

					if *deploy.Spec.Replicas != 2 {
						return fmt.Errorf("expected replicas 2, but got %d", *deploy.Spec.Replicas)
					}

					return nil
				}, timeout, time.Second).Should(gomega.Succeed())
			})

			ginkgo.By("delete the manifestworkreplicaset", func() {
				err := workClient.WorkV1alpha1().ManifestWorkReplicaSets(defaultNamespace).Delete(
					context.Background(), mwrsName, metav1.DeleteOptions{})
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				gomega.Eventually(func() error {
					_, err := workClient.WorkV1alpha1().ManifestWorkReplicaSets(defaultNamespace).Get(
						context.Background(), mwrsName, metav1.GetOptions{})
					if errors.IsNotFound(err) {
						return nil
					}
					if err != nil {
						return err
					}

					return fmt.Errorf("the %s/%s still exists", defaultNamespace, mwrsName)
				}, timeout, time.Second).Should(gomega.Succeed())
			})

			ginkgo.By("the manifests of the manifestworkreplicaset should be deleted from the managed cluster", func() {
				gomega.Eventually(func() error {
					_, err := spokeKubeClient.CoreV1().Namespaces().Get(
						context.Background(), manifestNamespace, metav1.GetOptions{})
					if errors.IsNotFound(err) {
						return nil
					}
					if err != nil {
						return err
					}

					return fmt.Errorf("the namespace %s still exists", manifestNamespace)
				}, timeout, time.Second).Should(gomega.Succeed())
			})
		})
	})
})

func newManifestWorkReplicaSet(name string, manifests []workv1.Manifest) *workv1alpha1.ManifestWorkReplicaSet {
	return &workv1alpha1.ManifestWorkReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: workv1alpha1.ManifestWorkReplicaSetSpec{
			PlacementRefs: []workv1alpha1.LocalPlacementReference{{Name: "all-clusters"}},
			ManifestWorkTemplate: workv1.ManifestWorkSpec{
				Workload: workv1.ManifestsTemplate{
					Manifests: manifests,
				},
			},
		},
	}
}

func newNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newDeployment(namespace, name string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "busybox"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "busybox"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: "quay.io/prometheus/busybox:latest",
							Args:  []string{"/bin/sh", "-c", "sleep 3600"},
						},
					},
				},
			},
		},
	}
}

func toManifest(object runtime.Object) workv1.Manifest {
	manifest := workv1.Manifest{}
	manifest.Object = object
	return manifest
}
