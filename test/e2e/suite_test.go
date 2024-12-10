package e2e_test

import (
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	workclientset "open-cluster-management.io/api/client/work/clientset/versioned"
)

const clusterName = "loopback"

const defaultNamespace = "default"

const timeout = 300 * time.Second

var (
	hubKubeClient   kubernetes.Interface
	spokeKubeClient kubernetes.Interface
	workClient      workclientset.Interface
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Maestro AddOn E2E Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	kubeconfig := os.Getenv("KUBECONFIG")

	hubConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	hubKubeClient, err = kubernetes.NewForConfig(hubConfig)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	workClient, err = workclientset.NewForConfig(hubConfig)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	spokeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	spokeKubeClient, err = kubernetes.NewForConfig(spokeConfig)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
})
