package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/matheusantonio/fc-ms-wallet/internal/database"
	"github.com/matheusantonio/fc-ms-wallet/internal/event"
	createaccount "github.com/matheusantonio/fc-ms-wallet/internal/usecase/create_account"
	createclient "github.com/matheusantonio/fc-ms-wallet/internal/usecase/create_client"
	createtransaction "github.com/matheusantonio/fc-ms-wallet/internal/usecase/create_transaction"
	"github.com/matheusantonio/fc-ms-wallet/internal/web"
	"github.com/matheusantonio/fc-ms-wallet/internal/web/webserver"
	"github.com/matheusantonio/fc-ms-wallet/pkg/events"
)

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", "root", "root", "localhost", "3306", "wallet"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	eventDispatcher := events.NewEventDispatcher()
	//eventDispatcher.Register("TransactionCreated", handler)
	transactionCreatedEvent := event.NewTransactionCreated()

	clientDb := database.NewClientDB(db)
	accountDb := database.NewAccountDB(db)
	transactionDb := database.NewTransactionDB(db)

	createClientUseCase := createclient.NewCreateClientUseCase(clientDb)
	createAccountUseCase := createaccount.NewCreateAccountUseCase(accountDb, clientDb)
	createTransactionUseCase := createtransaction.NewCreateTransactionUseCase(transactionDb, accountDb, eventDispatcher, transactionCreatedEvent)

	webServer := webserver.NewWebServer(":3000")

	clientHandler := web.NewWebClientHandler(*createClientUseCase)
	accountHandler := web.NewWebAccountHandler(*createAccountUseCase)
	transactionHandler := web.NewWebTransactionHandler(*createTransactionUseCase)

	webServer.AddHandler("/clients", clientHandler.CreateClient)
	webServer.AddHandler("/accounts", accountHandler.CreateAccount)
	webServer.AddHandler("/transactions", transactionHandler.CreateTransaction)

	webServer.Start()
}
