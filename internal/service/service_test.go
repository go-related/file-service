package service

import (
	"context"
	"github.com/go-related/fileservice/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestInsertions_HappyPath(t *testing.T) {
	testCases := map[string]struct {
		ctx             context.Context
		inputListResult []domain.Port //for simplicity i have put them on the same place normally we would have a different properties
	}{
		"InsertOneElement": {
			ctx: context.TODO(),
			inputListResult: []domain.Port{
				{
					Id: uuid.New().String(),
				},
			},
		},
	}
	//arrange
	mockRepository := new(MockRepository)
	server := NewPortService(mockRepository)
	t.Parallel()
	mockRepository.On("StartTransaction", mock.AnythingOfType("*context.cancelCtx"), true).Return(nil)
	mockRepository.On("CommitTransaction", mock.AnythingOfType("*context.cancelCtx")).Return(nil)
	mockRepository.On("AbortTransaction").Return()
	mockRepository.On("DoesTransactionExists").Return(true)
	mockRepository.On("AddOrUpdatePort", mock.Anything, mock.AnythingOfType("domain.Port")).Return(domain.Port{}, nil)
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			actualResult, err := server.AddOrUpdatePorts(test.ctx, test.inputListResult)
			assert.NoError(t, err)
			assert.Equal(t, test.inputListResult, actualResult)
			err = server.CommitTransaction(test.ctx)
			assert.NoError(t, err)
		})
	}
}

type MockRepository struct {
	mock.Mock
}

// AddOrUpdatePort mocks the AddOrUpdatePort method.
func (m *MockRepository) AddOrUpdatePort(ctx context.Context, item domain.Port) (domain.Port, error) {
	args := m.Called(ctx, item)
	return args.Get(0).(domain.Port), args.Error(1)
}

// AbortTransaction mocks the AbortTransaction method.
func (m *MockRepository) AbortTransaction() {
	m.Called()
}

// DoesTransactionExists mocks the DoesTransactionExists method.
func (m *MockRepository) DoesTransactionExists() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

// StartTransaction mocks the StartTransaction method.
func (m *MockRepository) StartTransaction(ctx context.Context, readOnly bool) error {
	args := m.Called(ctx, readOnly)
	return args.Error(0)
}

// CommitTransaction mocks the CommitTransaction method.
func (m *MockRepository) CommitTransaction(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
