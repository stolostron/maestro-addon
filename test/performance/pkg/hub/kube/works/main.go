package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	workclientset "open-cluster-management.io/api/client/work/clientset/versioned"

	"github.com/stolostron/maestro-addon/test/performance/pkg/common"
	"github.com/stolostron/maestro-addon/test/performance/pkg/util"
	"github.com/stolostron/maestro-addon/test/performance/pkg/workloads"
)

var (
	hubKubeConfigPath = flag.String("kubeconfig", "", "hub kubeconfig path")
	clusterBeginIndex = flag.Int("cluster-begin-index", 1, "Begin index of the clusters")
	clusterCounts     = flag.Int("cluster-counts", common.DEFAULT_CLUSTER_COUNTS, "Counts of the clusters")
	batches           = flag.Int("batches", 1, "")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer cancel()
		<-stopCh
	}()

	hubKubeConfig, err := clientcmd.BuildConfigFromFlags("", *hubKubeConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	workClient, err := workclientset.NewForConfig(hubKubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	works, err := workloads.ToACMManifestWorks()
	if err != nil {
		log.Fatal(err)
	}

	index := *clusterBeginIndex
	fmt.Printf("==== %s\n", time.Now().Format("2006-01-02 15:04:05"))
	for i := 0; i < *clusterCounts; i++ {
		clusterName := util.ClusterName(index)
		fmt.Printf("prepare works on cluster %s\n", clusterName)

		startTime := time.Now()
		for b := 0; b < *batches; b++ {
			for _, work := range works {
				newWork, err := workClient.WorkV1().ManifestWorks(clusterName).Create(
					context.Background(),
					workloads.CopyWork(b, work),
					metav1.CreateOptions{},
				)
				if errors.IsAlreadyExists(err) {
					continue
				}
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("work %s/%s is created\n", clusterName, newWork.Name)

				updatedWork := newWork.DeepCopy()
				updatedWork.Status = work.Status
				if _, err := workClient.WorkV1().ManifestWorks(clusterName).UpdateStatus(
					context.Background(),
					updatedWork,
					metav1.UpdateOptions{},
				); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("work %s/%s is updated\n", clusterName, newWork.Name)
			}
		}

		fmt.Printf("works (%d) are prepare on cluster %s, time=%dms\n",
			len(works)*(*batches), clusterName, util.UsedTime(startTime, time.Millisecond))
		index = index + 1
	}

	<-ctx.Done()
}
