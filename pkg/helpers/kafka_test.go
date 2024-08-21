package helpers

import (
	"context"
	"reflect"
	"testing"

	"github.com/stolostron/maestro-addon/pkg/helpers/mock"
)

func TestCreateKafkaTopics(t *testing.T) {
	cases := []struct {
		name           string
		topics         []string
		intiTopics     []string
		expectedTopics []string
	}{
		{
			name:           "create place holder topics",
			topics:         kafkaTopics(),
			intiTopics:     []string{},
			expectedTopics: kafkaTopics(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := mock.NewKafkaAdminMockClient(c.intiTopics...)
			if err := createKafkaTopics(context.Background(), client, c.topics...); err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(client.Topics(), c.expectedTopics) {
				t.Errorf("expected %v, but got %v", c.expectedTopics, client.Topics())
			}
		})
	}
}

func TestCreateKafkaACLs(t *testing.T) {
	cases := []struct {
		name         string
		topics       []string
		expectedACLs []string
	}{
		{
			name:         "create kafka cluster topics",
			topics:       kafkaTopics(),
			expectedACLs: append([]string{"*"}, kafkaTopics()...),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := mock.NewKafkaAdminMockClient()
			if err := createKafkaACLs(context.Background(), client, "cluster", c.topics...); err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(client.ACLs(), c.expectedACLs) {
				t.Errorf("expected %v, but got %v", c.expectedACLs, client.ACLs())
			}
		})
	}
}

func TestToKafkaPrincipal(t *testing.T) {
	expected := "User:CN=" +
		"system:open-cluster-management:cluster:cluster1:addon:maestro-addon:agent:maestro-addon-agent," +
		"O=system:authenticated+O=system:open-cluster-management:addon:maestro-addon+" +
		"O=system:open-cluster-management:cluster:cluster1:addon:maestro-addon"

	if toKafkaPrincipal("cluster1") != expected {
		t.Errorf("unexpected principal: %s", toKafkaPrincipal("cluster1"))
	}
}
