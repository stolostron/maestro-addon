package controllers

import (
	"context"
	"fmt"

	"github.com/openshift-online/maestro/pkg/api/openapi"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/stolostron/maestro-addon/pkg/helpers"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	clusterinformersv1 "open-cluster-management.io/api/client/cluster/informers/externalversions/cluster/v1"
	clusterlistersv1 "open-cluster-management.io/api/client/cluster/listers/cluster/v1"
)

const maestroServiceName = "maestro"

type ManagedClusterController struct {
	clusterLister    clusterlistersv1.ManagedClusterLister
	maestroAPIClient *openapi.APIClient
}

func NewManagedClusterController(clusterInformer clusterinformersv1.ManagedClusterInformer,
	recorder events.Recorder) factory.Controller {
	controller := &ManagedClusterController{
		clusterLister:    clusterInformer.Lister(),
		maestroAPIClient: helpers.NewMaestroAPIClient(fmt.Sprintf("http://%s:8000", maestroServiceName)),
	}

	return factory.New().
		WithInformersQueueKeysFunc(func(obj runtime.Object) []string {
			accessor, _ := meta.Accessor(obj)
			return []string{accessor.GetName()}
		}, clusterInformer.Informer()).
		WithSync(controller.sync).
		ToController("ManagedClusterController", recorder)
}

func (c *ManagedClusterController) sync(ctx context.Context, controllerContext factory.SyncContext) error {
	logger := klog.FromContext(ctx)

	managedClusterName := controllerContext.QueueKey()

	logger.Info("Reconciling ManagedCluster", "managedClusterName", managedClusterName)

	managedCluster, err := c.clusterLister.Get(managedClusterName)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if !managedCluster.DeletionTimestamp.IsZero() {
		// TODO delete this cluster in the maestro
		return nil
	}

	existed, err := c.findConsumerByName(ctx, managedClusterName)
	if err != nil {
		return err
	}

	if existed {
		return nil
	}

	// create a consumer in the maestro
	if _, _, err := c.maestroAPIClient.DefaultApi.ApiMaestroV1ConsumersPost(ctx).Consumer(openapi.Consumer{
		Name: openapi.PtrString(managedClusterName),
	}).Execute(); err != nil {
		return err
	}

	return nil
}

func (c *ManagedClusterController) findConsumerByName(ctx context.Context, managedClusterName string) (bool, error) {
	// TODO support to filer consumer by name in the maestro
	list, _, err := c.maestroAPIClient.DefaultApi.ApiMaestroV1ConsumersGet(ctx).Execute()
	if err != nil {
		return false, err
	}

	for _, consumer := range list.Items {
		if consumer.Name == &managedClusterName {
			return true, nil
		}
	}

	return false, nil
}
