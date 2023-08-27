package createtransaction

import (
	"context"
	"testing"

	"github.com/matheusantonio/fc-ms-wallet/internal/entity"
	"github.com/matheusantonio/fc-ms-wallet/internal/event"
	"github.com/matheusantonio/fc-ms-wallet/internal/usecase/mocks"
	"github.com/matheusantonio/fc-ms-wallet/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TransactionGatewayMock struct {
	mock.Mock
}

func (m *TransactionGatewayMock) Create(transaction *entity.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

type AccountGatewayMock struct {
	mock.Mock
}

func (m *AccountGatewayMock) Save(account *entity.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *AccountGatewayMock) FindByID(id string) (*entity.Account, error) {
	args := m.Called(id)
	return args.Get(0).(*entity.Account), args.Error(1)
}

func (m *AccountGatewayMock) UpdateBalance(account *entity.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func TestCreateTransactionUseCase_Execute(t *testing.T) {
	client1, _ := entity.NewClient("client1", "j@j1.com")
	account1 := entity.NewAccount(client1)
	account1.Credit(1000)

	client2, _ := entity.NewClient("client2", "j@j2.com")
	account2 := entity.NewAccount(client2)
	account2.Credit(1000)

	mockUow := &mocks.UowMock{}
	mockUow.On("Do", mock.Anything, mock.Anything).Return(nil)

	inputDto := CreateTransactionInputDto{
		AccountIDFrom: account1.ID,
		AccountIDTo:   account2.ID,
		Ammount:       100,
	}

	dispatcher := events.NewEventDispatcher()
	transactionEvent := event.NewTransactionCreated()
	balanceEvent := event.NewBalanceUpdated()
	ctx := context.Background()

	uc := NewCreateTransactionUseCase(mockUow, dispatcher, transactionEvent, balanceEvent)
	output, err := uc.Execute(ctx, inputDto)
	assert.Nil(t, err)
	assert.NotNil(t, output)
	mockUow.AssertExpectations(t)
	mockUow.AssertNumberOfCalls(t, "Do", 1)

}
