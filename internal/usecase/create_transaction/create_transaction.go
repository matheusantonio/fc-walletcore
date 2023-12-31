package createtransaction

import (
	"context"

	"github.com/matheusantonio/fc-ms-wallet/internal/entity"
	"github.com/matheusantonio/fc-ms-wallet/internal/gateway"
	"github.com/matheusantonio/fc-ms-wallet/pkg/events"
	"github.com/matheusantonio/fc-ms-wallet/pkg/uow"
)

type CreateTransactionInputDto struct {
	AccountIDFrom string  `json:"account_id_from"`
	AccountIDTo   string  `json:"account_id_to"`
	Ammount       float64 `json:"amount"`
}

type CreateTransactionOutputDto struct {
	ID            string
	AccountIDFrom string
	AccountIDTo   string
	Amount        float64
}

type BalanceUpdatedOutputDto struct {
	AccountIDFrom        string  `json:"account_id_from"`
	AccountIDTo          string  `json:"account_id_to"`
	BalanceAccountIDFrom float64 `json:"balance_account_id_from"`
	BalanceAccountIDTo   float64 `json:"balance_account_id_to"`
}

type CreateTransactionUseCase struct {
	Uow                uow.UowInterface
	EventDispatcher    events.EventDispatcherInterface
	TransactionCreated events.EventInterface
	BalanceUpdated     events.EventInterface
}

func NewCreateTransactionUseCase(
	uow uow.UowInterface,
	eventDispatcher events.EventDispatcherInterface,
	transactionCreated events.EventInterface,
	balanceUpdated events.EventInterface,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		Uow:                uow,
		EventDispatcher:    eventDispatcher,
		TransactionCreated: transactionCreated,
		BalanceUpdated:     balanceUpdated,
	}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInputDto) (*CreateTransactionOutputDto, error) {
	output := &CreateTransactionOutputDto{}
	balanceUpdated := &BalanceUpdatedOutputDto{}
	err := uc.Uow.Do(ctx, func(_ *uow.Uow) error {

		accountRepository := uc.getAccountRepository(ctx)
		transactionRepository := uc.getTransactionRepository(ctx)

		accountFrom, err := accountRepository.FindByID(input.AccountIDFrom)
		if err != nil {
			return err
		}

		accountTo, err := accountRepository.FindByID(input.AccountIDTo)
		if err != nil {
			return err
		}

		transaction, err := entity.NewTransaction(accountFrom, accountTo, input.Ammount)
		if err != nil {
			return err
		}

		err = accountRepository.UpdateBalance(accountTo)
		if err != nil {
			return err
		}

		err = accountRepository.UpdateBalance(accountFrom)
		if err != nil {
			return err
		}

		err = transactionRepository.Create(transaction)
		if err != nil {
			return err
		}

		output.ID = transaction.Id
		output.AccountIDFrom = transaction.AccountFrom.ID
		output.AccountIDTo = transaction.AccountTo.ID
		output.Amount = transaction.Amount

		balanceUpdated.AccountIDFrom = transaction.AccountFrom.ID
		balanceUpdated.AccountIDTo = transaction.AccountTo.ID
		balanceUpdated.BalanceAccountIDFrom = transaction.AccountFrom.Balance
		balanceUpdated.BalanceAccountIDTo = transaction.AccountTo.Balance

		return nil

	})
	if err != nil {
		return nil, err
	}

	uc.TransactionCreated.SetPayload(output)
	uc.EventDispatcher.Dispatch(uc.TransactionCreated)

	uc.BalanceUpdated.SetPayload(balanceUpdated)
	uc.EventDispatcher.Dispatch(uc.BalanceUpdated)

	return output, nil

}

func (uc *CreateTransactionUseCase) getAccountRepository(ctx context.Context) gateway.AccountGateway {
	repository, err := uc.Uow.GetRepository(ctx, "AccountDB")
	if err != nil {
		panic(err)
	}
	return repository.(gateway.AccountGateway)
}

func (uc *CreateTransactionUseCase) getTransactionRepository(ctx context.Context) gateway.TransactionGateway {
	repository, err := uc.Uow.GetRepository(ctx, "TransactionDB")
	if err != nil {
		panic(err)
	}
	return repository.(gateway.TransactionGateway)
}
