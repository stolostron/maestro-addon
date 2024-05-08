package helpers

import (
	"context"
	"testing"

	"github.com/stolostron/maestro-addon/pkg/helpers/mock"
)

func TestFindConsumerByName(t *testing.T) {
	maestroServer := mock.NewMaestroMockServer()
	maestroServer.Start()
	defer maestroServer.Stop()

	cases := []struct {
		name     string
		consumer string
		expected bool
	}{
		{
			name:     "find an existed consumer",
			consumer: mock.Consumer,
			expected: true,
		},
		{
			name:     "find a nonexistent consumer",
			consumer: "cluster1",
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := FindConsumerByName(
				context.Background(), NewMaestroAPIClient(maestroServer.URL()), c.consumer)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != c.expected {
				t.Errorf("unexpected result: %t", result)
			}
		})
	}
}

func TestCreateConsumer(t *testing.T) {
	maestroServer := mock.NewMaestroMockServer()
	maestroServer.Start()
	defer maestroServer.Stop()

	if err := CreateConsumer(context.Background(), NewMaestroAPIClient(maestroServer.URL()), "test"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

}
