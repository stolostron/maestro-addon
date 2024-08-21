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
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options/grpc"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/source/codec"

	"github.com/stolostron/maestro-addon/test/performance/pkg/common"
	"github.com/stolostron/maestro-addon/test/performance/pkg/hub/maestro/store"
	"github.com/stolostron/maestro-addon/test/performance/pkg/util"
	"github.com/stolostron/maestro-addon/test/performance/pkg/workloads"
)

const sourceID = "maestro"

var (
	grpcServiceAddress = flag.String("grpc-service-address", "127.0.0.1:8090", "")
	clusterBeginIndex  = flag.Int("cluster-begin-index", 1, "Begin index of the clusters")
	clusterCounts      = flag.Int("cluster-counts", common.DEFAULT_AGENT_COUNTS, "Counts of the clusters")
	batches            = flag.Int("batches", 5, "")
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

	creator, err := work.NewClientHolderBuilder(&grpc.GRPCOptions{URL: *grpcServiceAddress}).
		WithClientID(fmt.Sprintf("%s-client", sourceID)).
		WithSourceID(sourceID).
		WithCodecs(codec.NewManifestBundleCodec()).
		WithWorkClientWatcherStore(store.NewCreateOnlyWatcherStore()).
		WithResyncEnabled(false).
		NewSourceClientHolder(ctx)
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

		for _, work := range works {
			startTime := time.Now()

			newWork := workloads.CopyWork(*batches, work)
			if _, err := creator.ManifestWorks(clusterName).Create(
				ctx,
				newWork,
				metav1.CreateOptions{},
			); err != nil {
				if !errors.IsAlreadyExists(err) {
					log.Fatal(err)
				}
			}
			fmt.Printf("the work %s/%s is created, time=%dms\n",
				clusterName, newWork.Name, util.UsedTime(startTime, time.Millisecond))
		}

		if i%20 == 0 {
			time.Sleep(2 * time.Second)
		}

		index = index + 1
	}

	<-ctx.Done()
}
