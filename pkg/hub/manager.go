package hub

import (
	"context"
	"time"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"

	addonv1alpha1client "open-cluster-management.io/api/client/addon/clientset/versioned"
	addoninformers "open-cluster-management.io/api/client/addon/informers/externalversions"
	clusterclientset "open-cluster-management.io/api/client/cluster/clientset/versioned"
	clusterinformers "open-cluster-management.io/api/client/cluster/informers/externalversions"

	"github.com/stolostron/maestro-addon/pkg/common"
	"github.com/stolostron/maestro-addon/pkg/helpers"
	"github.com/stolostron/maestro-addon/pkg/hub/controllers"
)

const (
	defaultAddOnManagerNamespace = "open-cluster-management-hub"
	defaultAgentNamespace        = "open-cluster-management-agent"
)

// MaestroAddOnManagerOptions defines the flags for maestro-addon hub manager
type MaestroAddOnManagerOptions struct {
	AddOnManagerNamespace          string
	AgentNamespace                 string
	UseCustomizedMessageQueueCerts bool
}

func NewMaestroAddOnManagerOptions() *MaestroAddOnManagerOptions {
	return &MaestroAddOnManagerOptions{
		AddOnManagerNamespace:          defaultAddOnManagerNamespace,
		AgentNamespace:                 defaultAgentNamespace,
		UseCustomizedMessageQueueCerts: false,
	}
}

// AddFlags register and binds the default flags
func (o *MaestroAddOnManagerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.AddOnManagerNamespace, "addon-manager-namespace", o.AddOnManagerNamespace,
		"The AddOnManager namespace")
	fs.StringVar(&o.AgentNamespace, "agent-namespace", o.AgentNamespace,
		"The agent namespace")
	fs.BoolVar(&o.UseCustomizedMessageQueueCerts, "use-customized-mq-certs", o.UseCustomizedMessageQueueCerts,
		"If false, the manager will generate the message queue certs")
}

func (o *MaestroAddOnManagerOptions) RunHubManager(ctx context.Context, controllerContext *controllercmd.ControllerContext) error {
	kubeClient, err := kubernetes.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}

	clusterClient, err := clusterclientset.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}

	addOnClient, err := addonv1alpha1client.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}

	// the message queue certs are not provided by user, prepare them
	if !o.UseCustomizedMessageQueueCerts {
		// TODO need maintain the mq certs if we create them
		if err := helpers.PrepareCerts(
			ctx, kubeClient, controllerContext.OperatorNamespace, o.AddOnManagerNamespace); err != nil {
			return err
		}
	}

	clusterInformers := clusterinformers.NewSharedInformerFactory(clusterClient, 30*time.Minute)
	addonInformers := addoninformers.NewSharedInformerFactoryWithOptions(
		addOnClient,
		30*time.Minute,
		addoninformers.WithTweakListOptions(
			func(options *metav1.ListOptions) {
				options.FieldSelector = fields.OneTermEqualSelector("metadata.name", common.AddOnTemplateName).String()
			}),
	)

	addonTemplateController := controllers.NewAddOnTemplateController(
		kubeClient,
		addOnClient,
		addonInformers.Addon().V1alpha1().AddOnTemplates(),
		controllerContext.OperatorNamespace,
		o.AgentNamespace,
		controllerContext.EventRecorder,
	)
	managedClusterController := controllers.NewManagedClusterController(
		clusterInformers.Cluster().V1().ManagedClusters(),
		controllerContext.EventRecorder,
	)

	go clusterInformers.Start(ctx.Done())
	go addonInformers.Start(ctx.Done())

	go managedClusterController.Run(ctx, 1)
	go addonTemplateController.Run(ctx, 1)

	<-ctx.Done()
	return nil
}
