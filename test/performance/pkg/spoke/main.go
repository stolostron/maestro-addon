package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/cache"
	workinformers "open-cluster-management.io/api/client/work/informers/externalversions"
	workv1 "open-cluster-management.io/api/work/v1"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options/kafka"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/agent/codec"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/store"

	"github.com/stolostron/maestro-addon/test/performance/pkg/common"
	"github.com/stolostron/maestro-addon/test/performance/pkg/util"
	"github.com/stolostron/maestro-addon/test/performance/pkg/workloads"
)

var (
	workDir           = flag.String("work-dir", "", "")
	clusterBeginIndex = flag.Int("cluster-begin-index", 1, "Begin index of the clusters")
	clusterCounts     = flag.Int("cluster-counts", common.DEFAULT_AGENT_COUNTS, "Counts of the clusters")
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

	works, err := workloads.ToACMManifestWorks()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("==== %s\n", time.Now().Format("2006-01-02 15:04:05"))
	index := *clusterBeginIndex
	for i := 0; i < *clusterCounts; i++ {
		clusterName := util.ClusterName(index)
		if err := startWorkAgent(ctx, clusterName, works); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("agent %s is started\n", clusterName)
		if i > 99 && i%100 == 0 {
			time.Sleep(10 * time.Second)
		}
		index = index + 1
	}

	<-ctx.Done()
}

func startWorkAgent(ctx context.Context, clusterName string, works map[string]*workv1.ManifestWork) error {
	watcherStore := store.NewAgentInformerWatcherStore()

	config, err := kafka.BuildKafkaOptionsFromFlags(
		filepath.Join(*workDir, "config", fmt.Sprintf("client-%s.config", clusterName)))
	if err != nil {
		return err
	}
	clientHolder, err := work.NewClientHolderBuilder(config).
		WithClientID(clusterName + "-" + rand.String(5)).
		WithClusterName(clusterName).
		WithCodecs(codec.NewManifestBundleCodec()).
		WithWorkClientWatcherStore(watcherStore).
		NewAgentClientHolder(ctx)
	if err != nil {
		return err
	}

	factory := workinformers.NewSharedInformerFactoryWithOptions(
		clientHolder.WorkInterface(),
		5*time.Minute,
		workinformers.WithNamespace(clusterName),
	)
	informer := factory.Work().V1().ManifestWorks().Informer()
	if _, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			work, ok := obj.(*workv1.ManifestWork)
			if !ok {
				fmt.Printf("unknown type %v\n", obj)
			}
			fmt.Printf("work %s/%s is watched\n", clusterName, work.Name)

			patchs, err := patch(work, works)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}

			if _, err := clientHolder.ManifestWorks(clusterName).Patch(
				ctx,
				work.Name,
				types.MergePatchType,
				patchs,
				metav1.PatchOptions{},
				"status",
			); err != nil {
				fmt.Printf("error: %v\n", err)
			}

			fmt.Printf("work %s/%s is updated\n", clusterName, work.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	}); err != nil {
		log.Fatal(err)
	}
	watcherStore.SetInformer(informer)

	go informer.Run(ctx.Done())

	return nil
}

func patch(work *workv1.ManifestWork, works map[string]*workv1.ManifestWork) ([]byte, error) {
	oldData, err := json.Marshal(work)
	if err != nil {
		return nil, err
	}

	newWork := work.DeepCopy()
	for name, twork := range works {
		if strings.HasPrefix(work.Name, name) {
			fmt.Printf("status for %s\n", work.Name)
			newWork.Status = twork.Status
			break
		}
	}

	newData, err := json.Marshal(newWork)
	if err != nil {
		return nil, err
	}

	return jsonpatch.CreateMergePatch(oldData, newData)
}
