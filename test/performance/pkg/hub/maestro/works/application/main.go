package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

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
)

func main() {
	flag.Parse()

	creator, err := work.NewClientHolderBuilder(&grpc.GRPCOptions{URL: *grpcServiceAddress}).
		WithClientID(fmt.Sprintf("%s-client", sourceID)).
		WithSourceID(sourceID).
		WithCodecs(codec.NewManifestBundleCodec()).
		WithWorkClientWatcherStore(store.NewCreateOnlyWatcherStore()).
		WithResyncEnabled(false).
		NewSourceClientHolder(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	index := *clusterBeginIndex
	fmt.Printf("==== %s\n", time.Now().Format("2006-01-02 15:04:05"))
	for i := 0; i < *clusterCounts; i++ {
		clusterName := util.ClusterName(index)

		works, err := workloads.ToGuestBookWorks(clusterName, 50)
		if err != nil {
			log.Fatal(err)
		}

		for j, work := range works {
			startTime := time.Now()
			_, err := creator.ManifestWorks(clusterName).Create(
				context.Background(),
				work,
				metav1.CreateOptions{},
			)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("the work %s/%s is created, time=%dms\n",
					clusterName, work.Name, util.UsedTime(startTime, time.Millisecond))
			}

			if j%10 == 0 {
				time.Sleep(time.Second)
			}
		}

		time.Sleep(2 * time.Second)

		index = index + 1
	}
}
