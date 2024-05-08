package helpers

import (
	"context"
	"fmt"
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

func FindConsumerByName(ctx context.Context, client *openapi.APIClient, consumerName string) (bool, error) {
	list, _, err := client.DefaultApi.ApiMaestroV1ConsumersGet(ctx).
		Search(fmt.Sprintf("name = '%s'", consumerName)).
		Execute()
	if err != nil {
		return false, err
	}

	for _, consumer := range list.Items {
		if *consumer.Name == consumerName {
			return true, nil
		}
	}

	return false, nil
}

func CreateConsumer(ctx context.Context, client *openapi.APIClient, consumerName string) error {
	_, _, err := client.DefaultApi.ApiMaestroV1ConsumersPost(ctx).
		Consumer(openapi.Consumer{Name: openapi.PtrString(consumerName)}).
		Execute()
	return err
}
