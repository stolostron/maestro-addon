package helpers

import (
	"net/http"
	"os"
	"time"

	"github.com/openshift-online/maestro/pkg/api/openapi"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"

	workv1 "open-cluster-management.io/api/work/v1"
)

const defaultComponentNamespace = "maestro"

func GetComponentNamespace() string {
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err == nil {
		return string(nsBytes)
	}

	return defaultComponentNamespace
}

func NewMaestroAPIClient(maestroServerAddress string) *openapi.APIClient {
	cfg := &openapi.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "OpenAPI-Generator/1.0.0/go",
		Debug:         false,
		Servers: openapi.ServerConfigurations{
			{
				URL:         maestroServerAddress,
				Description: "current domain",
			},
		},
		OperationServers: map[string]openapi.ServerConfigurations{},
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	return openapi.NewAPIClient(cfg)
}

func ToManifest(object runtime.Object) workv1.Manifest {
	manifest := workv1.Manifest{}
	manifest.Object = object
	return manifest
}

func ManifestsEqual(oldManifests, newManifests []workv1.Manifest) bool {
	if len(oldManifests) != len(newManifests) {
		return false
	}

	for i := range oldManifests {
		if !equality.Semantic.DeepEqual(oldManifests[i].Raw, newManifests[i].Raw) {
			return false
		}
	}
	return true
}
