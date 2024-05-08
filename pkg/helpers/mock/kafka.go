package mock

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaAdminMockClient struct {
	topics kafka.DescribeTopicsResult
	acls   *kafka.DescribeACLsResult
}

func NewKafkaAdminMockClient(initTopics ...string) *KafkaAdminMockClient {
	topics := []kafka.TopicDescription{}
	for _, topic := range initTopics {
		topics = append(topics, kafka.TopicDescription{
			Name:  topic,
			Error: kafka.NewError(kafka.ErrNoError, "", false),
		})
	}

	return &KafkaAdminMockClient{
		topics: kafka.DescribeTopicsResult{
			TopicDescriptions: topics,
		},
		acls: &kafka.DescribeACLsResult{
			ACLBindings: kafka.ACLBindings{},
			Error:       kafka.NewError(kafka.ErrNoError, "", false),
		},
	}
}

func (m *KafkaAdminMockClient) DescribeTopics(ctx context.Context, topics kafka.TopicCollection,
	options ...kafka.DescribeTopicsAdminOption) (result kafka.DescribeTopicsResult, err error) {
	return m.topics, nil
}

func (m *KafkaAdminMockClient) DescribeACLs(ctx context.Context, aclBindingFilter kafka.ACLBindingFilter,
	options ...kafka.DescribeACLsAdminOption) (result *kafka.DescribeACLsResult, err error) {
	return m.acls, nil
}

func (m *KafkaAdminMockClient) CreateTopics(ctx context.Context, topics []kafka.TopicSpecification,
	options ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error) {
	for _, topic := range topics {
		m.topics.TopicDescriptions = append(m.topics.TopicDescriptions, kafka.TopicDescription{
			Name:  topic.Topic,
			Error: kafka.NewError(kafka.ErrNoError, "", false),
		})

		result = append(result, kafka.TopicResult{
			Topic: topic.Topic,
			Error: kafka.NewError(kafka.ErrNoError, "", false),
		})
	}
	return result, nil
}

func (m *KafkaAdminMockClient) CreateACLs(ctx context.Context, aclBindings kafka.ACLBindings,
	options ...kafka.CreateACLsAdminOption) (result []kafka.CreateACLResult, err error) {
	for _, binding := range aclBindings {
		m.acls.ACLBindings = append(m.acls.ACLBindings, binding)
		result = append(result, kafka.CreateACLResult{
			Error: kafka.NewError(kafka.ErrNoError, "", false),
		})
	}
	return result, nil
}

func (m *KafkaAdminMockClient) Topics() []string {
	topics := []string{}
	for _, topic := range m.topics.TopicDescriptions {
		topics = append(topics, topic.Name)
	}
	return topics
}

func (m *KafkaAdminMockClient) ACLs() []string {
	acls := []string{}
	for _, acl := range m.acls.ACLBindings {
		acls = append(acls, acl.Name)
	}
	return acls
}
