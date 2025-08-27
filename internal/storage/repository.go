package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"golang-server/internal/model"
	"golang-server/internal/storage/postgres"
	"math/big"
)

type TransactionRepository struct {
	db *postgres.PgDB
}

type WalletRepository struct {
	db *postgres.PgDB
}

func NewTransactionRepository(db *postgres.PgDB) *TransactionRepository {
	return &TransactionRepository{db: db} //todo во всех конструктарах мы передаём db, что странно (отрефакторить)
}

func NewWalletRepository(db *postgres.PgDB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (wr *WalletRepository) GetBalance(walletId uuid.UUID) (*big.Float, error) {
	db := wr.db

	getBalanceQuery := "SELECT balance FROM wallets WHERE id = $1;"
	var balance float64
	err := db.QueryRow(getBalanceQuery, walletId).Scan(&balance) //todo падает, если json (body) пуст
	if err != nil {
		return nil, err
	}
	return big.NewFloat(balance), nil
}

// todo реализовать
// todo дописать корректный уровень изоляции
func (tr *TransactionRepository) SendMoney(data model.TransferMoneyRequest) (*uuid.UUID, error) { //todo продумать ответ сервера
	db := tr.db
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	sendQuery := "INSERT INTO transactions (id, from_wallet, to_wallet, amount, transfer_date) VALUES ($1, $2, $3, $4, NOW())"
	transactionId := uuid.New()
	_, err = tx.Exec(sendQuery, transactionId, data.From, data.To, data.Amount)
	if err != nil {
		return nil, err
	}

	updateQuery := "UPDATE wallets SET balance = balance + $1, date_update = NOW() WHERE id = $2"
	_, err = tx.Exec(updateQuery, data.Amount, data.To)
	if err != nil {
		return nil, err
	}

	updateQuery = "UPDATE wallets SET balance = balance - $1, date_update = NOW() WHERE id = $2"
	_, err = tx.Exec(updateQuery, data.Amount, data.From)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &transactionId, nil
}

func (tr *TransactionRepository) GetLastTransactions(numberOfTx int) ([]model.Transaction, error) {
	query :=
		`SELECT id, from_wallet, to_wallet, amount, transfer_date
	FROM transactions
	ORDER BY transfer_date
	LIMIT $1;`

	rows, err := tr.db.Query(query, numberOfTx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]model.Transaction, 0, numberOfTx)

	for rows.Next() {
		var transaction model.Transaction
		var amountStr string

		err := rows.Scan(
			&transaction.Id,
			&transaction.To,
			&transaction.From,
			&amountStr,
			&transaction.TransferDate,
		)
		if err != nil {
			return nil, err
		}

		amount := new(big.Float)
		_, ok := amount.SetString(amountStr)
		if !ok {
			return nil, fmt.Errorf("failed to parse amount: %s", amountStr)
		}
		transaction.Amount = *amount

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (wr *WalletRepository) GetWallet(walletId uuid.UUID) (*model.Wallet, error) {
	query :=
		`SELECT
	w.id,
		w.balance,
		w.date_update
	FROM wallets w
	WHERE w.id = $1;`

	var wallet model.Wallet
	var balanceStr string

	err := wr.db.QueryRow(query, walletId).Scan(&wallet.Id, &balanceStr, &wallet.DateUpdate)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Парсим balanceStr в big.Float
	balance := new(big.Float)
	if _, ok := balance.SetString(balanceStr); !ok {
		return nil, fmt.Errorf("failed to parse balance: %s", balanceStr)
	}
	wallet.Balance = *balance

	return &wallet, nil
}
