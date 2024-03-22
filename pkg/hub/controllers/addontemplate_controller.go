package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/stolostron/maestro-addon/pkg/common"
	"github.com/stolostron/maestro-addon/pkg/helpers"

	addonv1alpha1client "open-cluster-management.io/api/client/addon/clientset/versioned"
	addonv1alpha1informers "open-cluster-management.io/api/client/addon/informers/externalversions/addon/v1alpha1"
	addonv1alpha1listers "open-cluster-management.io/api/client/addon/listers/addon/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"
)

const mqCertsConfigMapName = "maestro-mq-ca"

type AddOnTemplateController struct {
	kubeClient          kubernetes.Interface
	addonClient         addonv1alpha1client.Interface
	addOnTemplateLister addonv1alpha1listers.AddOnTemplateLister
	namespace           string
	agentNamespace      string
}

func NewAddOnTemplateController(kubeClient kubernetes.Interface,
	addonClient addonv1alpha1client.Interface,
	addOnTemplateInformer addonv1alpha1informers.AddOnTemplateInformer,
	namespace string,
	agentNamespace string,
	recorder events.Recorder) factory.Controller {
	c := &AddOnTemplateController{
		kubeClient:          kubeClient,
		addonClient:         addonClient,
		addOnTemplateLister: addOnTemplateInformer.Lister(),
		namespace:           namespace,
		agentNamespace:      agentNamespace,
	}
	return factory.New().
		WithInformers(addOnTemplateInformer.Informer()).
		WithSync(c.sync).
		ToController("MessageQueueCertsController", recorder)
}

func (c *AddOnTemplateController) sync(ctx context.Context, controllerContext factory.SyncContext) error {
	logger := klog.FromContext(ctx)
	logger.Info("Reconciling AddOnTemplate", "addOnTemplateName", common.AddOnTemplateName)

	template, err := c.addOnTemplateLister.Get(common.AddOnTemplateName)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	secret, err := c.kubeClient.CoreV1().Secrets(c.namespace).Get(ctx, common.MessageQueueCertsSecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to find secret %s/%s: %w", c.namespace, common.AddOnTemplateName, err)
	}

	if _, ok := secret.Data[common.MessageQueueCAKey]; !ok {
		return fmt.Errorf("no `%s` in the secret %s/%s", common.MessageQueueCAKey, c.namespace, common.AddOnTemplateName)
	}

	manifests, err := newManifests(template.Spec.AgentSpec.Workload.Manifests, c.agentNamespace, secret)
	if err != nil {
		return err
	}

	if helpers.ManifestsEqual(template.Spec.AgentSpec.Workload.Manifests, manifests) {
		return nil
	}

	newTemplate := template.DeepCopy()
	newTemplate.Spec.AgentSpec.Workload.Manifests = manifests
	if _, err := c.addonClient.AddonV1alpha1().AddOnTemplates().Update(ctx, newTemplate, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func newManifests(manifests []workv1.Manifest, agentNamespace string, secret *corev1.Secret) ([]workv1.Manifest, error) {
	newManifests := []workv1.Manifest{}
	for _, manifest := range manifests {
		// parse the required and set resource meta
		required := &unstructured.Unstructured{}
		if err := required.UnmarshalJSON(manifest.Raw); err != nil {
			return nil, err
		}
		if required.GetName() != mqCertsConfigMapName {
			newManifests = append(newManifests, manifest)
		}
	}

	return append(newManifests, helpers.ToManifest(newConfigmap(agentNamespace, secret))), nil
}

func newConfigmap(namespace string, secret *corev1.Secret) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      mqCertsConfigMapName,
		},
		Data: map[string]string{
			common.MessageQueueCAKey: string(secret.Data[common.MessageQueueCAKey]),
		},
	}
}
