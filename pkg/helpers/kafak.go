package helpers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// CreteKafkaPlaceholderTopics creates placeholder topics.
// This avoids unknown topic error when subscribing to wildcard topics
func CreteKafkaPlaceholderTopics(ctx context.Context, config *kafka.ConfigMap, sourceID string) error {
	client, err := kafka.NewAdminClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	return createKafkaTopics(ctx,
		client,
		fmt.Sprintf("sourceevents.%s.agent", sourceID),
		fmt.Sprintf("sourcebroadcast.%s", sourceID),
		fmt.Sprintf("agentevents.%s.agent", sourceID),
		"agentbroadcast.agent",
	)
}

func CreateKafkaTopicsWithACLs(ctx context.Context, config *kafka.ConfigMap, sourceID, clusterName string) error {
	adminClient, err := kafka.NewAdminClient(config)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	// each cluster has four topics
	sourceEventsTopic := fmt.Sprintf("sourceevents.%s.%s", sourceID, clusterName)
	sourceBroadcastTopic := fmt.Sprintf("sourcebroadcast.%s", sourceID)
	agentEventsTopic := fmt.Sprintf("agentevents.%s.%s", sourceID, clusterName)
	agentBroadcastTopic := fmt.Sprintf("agentbroadcast.%s", clusterName)

	if err := createKafkaTopics(ctx, adminClient, sourceEventsTopic, sourceBroadcastTopic,
		agentEventsTopic, agentBroadcastTopic); err != nil {
		return err
	}

	// TODO: common name
	commonName := fmt.Sprintf("maestro-kafka-%s", clusterName)
	if err := createKafkaACLs(ctx, adminClient, commonName, sourceEventsTopic, sourceBroadcastTopic,
		agentEventsTopic, agentBroadcastTopic); err != nil {
		return err
	}

	return nil
}

func createKafkaTopics(ctx context.Context, adminClient *kafka.AdminClient, newTopics ...string) error {
	logger := klog.FromContext(ctx)

	topics, err := adminClient.DescribeTopics(ctx, kafka.NewTopicCollectionOfTopicNames(newTopics))
	if err != nil {
		return err
	}

	topicSpecs := []kafka.TopicSpecification{}
	for _, topic := range newTopics {
		if hasTopic(topics.TopicDescriptions, topic) {
			logger.V(4).Info(fmt.Sprintf("topic %s already exists", topic))
			continue
		}

		topicSpecs = append(topicSpecs, kafka.TopicSpecification{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	if len(topicSpecs) == 0 {
		return nil
	}

	results, err := adminClient.CreateTopics(ctx, topicSpecs)
	if err != nil {
		return err
	}

	errs := []error{}
	for _, r := range results {
		if r.Error.Code() == kafka.ErrNoError {
			logger.V(4).Info(fmt.Sprintf("topic %s created successfully", r.Topic))
			continue
		}

		errs = append(errs, fmt.Errorf("failed to create topic %s, %s", r.Topic, r.Error.String()))
	}

	return errors.NewAggregate(errs)
}

func createKafkaACLs(ctx context.Context, adminClient *kafka.AdminClient, commonName string, topics ...string) error {
	logger := klog.FromContext(ctx)

	expectedACLBindings := []kafka.ACLBinding{{
		Type:                kafka.ResourceGroup,
		Name:                "*",
		ResourcePatternType: kafka.ResourcePatternTypeLiteral,
		Principal:           fmt.Sprintf("User:CN=%s", commonName),
		Host:                "*",
		Operation:           kafka.ACLOperationAll,
		PermissionType:      kafka.ACLPermissionTypeAllow,
	}}

	for _, topic := range topics {
		expectedACLBindings = append(expectedACLBindings, kafka.ACLBinding{
			Type:                kafka.ResourceTopic,
			Name:                topic,
			ResourcePatternType: kafka.ResourcePatternTypeLiteral,
			Principal:           fmt.Sprintf("User:CN=%s", commonName),
			Host:                "*",
			Operation:           kafka.ACLOperationAll,
			PermissionType:      kafka.ACLPermissionTypeAllow,
		})
	}

	aclBindings := []kafka.ACLBinding{}
	for _, acl := range expectedACLBindings {
		result, err := adminClient.DescribeACLs(ctx, acl)
		if err != nil {
			return err
		}

		if hasACL(result, acl) {
			logger.V(4).Info(fmt.Sprintf("acl %s/%s already exists for %s\n", acl.Type, acl.Name, acl.Principal))
			continue
		}

		aclBindings = append(aclBindings, acl)
	}

	if len(aclBindings) == 0 {
		return nil
	}

	results, err := adminClient.CreateACLs(ctx, aclBindings)
	if err != nil {
		return err
	}

	errs := []error{}
	for _, r := range results {
		if r.Error.Code() != kafka.ErrNoError {
			errs = append(errs, fmt.Errorf("failed to create acl %s", r.Error.String()))
		}
	}
	if len(errs) == 0 {
		logger.V(4).Info(fmt.Sprintf("acls is created successfully for agent %s\n", commonName))
	}

	return errors.NewAggregate(errs)
}

func hasTopic(topics []kafka.TopicDescription, topic string) bool {
	for _, t := range topics {
		if t.Error.Code() == kafka.ErrNoError && t.Name == topic {
			return true
		}
	}

	return false
}

func hasACL(acls *kafka.DescribeACLsResult, binding kafka.ACLBinding) bool {
	if acls.Error.Code() == kafka.ErrNoError {
		for _, a := range acls.ACLBindings {
			if a.Name == binding.Name {
				return true
			}
		}
	}
	return false
}
