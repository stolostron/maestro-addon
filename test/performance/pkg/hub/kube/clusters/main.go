package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/stolostron/maestro-addon/test/performance/pkg/common"
	"github.com/stolostron/maestro-addon/test/performance/pkg/util"
)

var (
	hubKubeConfigPath = flag.String("kubeconfig", "", "hub kubeconfig path")
	beginIndex        = flag.Int("begin-index", 1, "Begin index of the clusters")
	counts            = flag.Int("counts", common.DEFAULT_CLUSTER_COUNTS, "Counts of the clusters")
)

func main() {
	flag.Parse()

	hubKubeConfig, err := clientcmd.BuildConfigFromFlags("", *hubKubeConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	hubKubeClient, err := kubernetes.NewForConfig(hubKubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	index := *beginIndex
	startTime := time.Now()
	for i := 0; i < *counts; i++ {
		clusterName := util.ClusterName(index)

		startTime := time.Now()
		if _, err := hubKubeClient.CoreV1().Namespaces().Create(
			context.Background(),
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name: clusterName,
				Labels: map[string]string{
					"maestro.performance.test": "acm",
				},
			}},
			metav1.CreateOptions{},
		); err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Fatal(err)
			}
		}
		index = index + 1
		fmt.Printf("cluster namespace %s is created, time=%dms\n", clusterName, util.UsedTime(startTime, time.Millisecond))
	}
	fmt.Printf("Clusters (%d) are created, time=%dms\n", *counts, util.UsedTime(startTime, time.Millisecond))
}
