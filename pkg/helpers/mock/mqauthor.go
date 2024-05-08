package mock

import "context"

type MockMessageQueueAuthzCreator struct {
	clusterName string
}

func NewMockMessageQueueAuthzCreator() *MockMessageQueueAuthzCreator {
	return &MockMessageQueueAuthzCreator{}
}

func (a *MockMessageQueueAuthzCreator) CreateAuthorizations(ctx context.Context, clusterName string) error {
	a.clusterName = clusterName
	return nil
}

func (a *MockMessageQueueAuthzCreator) DeleteAuthorizations(ctx context.Context, clusterName string) error {
	return nil
}

func (a *MockMessageQueueAuthzCreator) ClusterName() string {
	return a.clusterName
}
