package helpers

import (
	"net/http"
	"time"

	"github.com/openshift-online/maestro/pkg/api/openapi"
)

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
