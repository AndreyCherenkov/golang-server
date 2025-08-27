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

// TransactionRepository управляет транзакциями между кошельками.
type TransactionRepository struct {
	db *postgres.PgDB
}

// WalletRepository управляет данными кошельков.
type WalletRepository struct {
	db *postgres.PgDB
}

// NewTransactionRepository создаёт новый репозиторий транзакций.
func NewTransactionRepository(db *postgres.PgDB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// NewWalletRepository создаёт новый репозиторий кошельков.
func NewWalletRepository(db *postgres.PgDB) *WalletRepository {
	return &WalletRepository{db: db}
}

// GetBalance возвращает текущий баланс кошелька по его ID.
// Параметры:
//   - walletId: идентификатор кошелька.
//
// Возвращает:
//   - *big.Float: баланс кошелька.
//   - error: ошибку при выполнении запроса.
func (wr *WalletRepository) GetBalance(walletId uuid.UUID) (*big.Float, error) {
	db := wr.db

	// Начинаем транзакцию с Repeatable Read
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback() // безопасно вызвать Rollback даже после Commit
	}()

	getBalanceQuery := "SELECT balance FROM wallets WHERE id = $1;"
	var balance float64
	err = tx.QueryRow(getBalanceQuery, walletId).Scan(&balance)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return big.NewFloat(balance), nil
}

// SendMoney переводит деньги между кошельками в рамках транзакции.
// Уровень изоляции: Serializable для предотвращения проблем с конкурентным доступом (Race Conditions, Phantom Reads).
// Параметры:
//   - data: структура TransferMoneyRequest с информацией о переводе.
//
// Возвращает:
//   - *uuid.UUID: ID созданной транзакции.
//   - error: ошибку при выполнении транзакции.
func (tr *TransactionRepository) SendMoney(data model.TransferMoneyRequest) (*uuid.UUID, error) {
	db := tr.db
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Высокий уровень изоляции предотвращает конкурентные конфликты
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			panic(p)
		} else if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
		}
	}()

	// Вставка новой транзакции
	sendQuery := "INSERT INTO transactions (id, from_wallet, to_wallet, amount, transfer_date) VALUES ($1, $2, $3, $4, NOW())"
	transactionId := uuid.New()
	_, err = tx.Exec(sendQuery, transactionId, data.From, data.To, data.Amount)
	if err != nil {
		return nil, err
	}

	// Увеличение баланса получателя
	updateQuery := "UPDATE wallets SET balance = balance + $1, date_update = NOW() WHERE id = $2"
	_, err = tx.Exec(updateQuery, data.Amount, data.To)
	if err != nil {
		return nil, err
	}

	// Уменьшение баланса отправителя
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

// GetLastTransactions возвращает последние N транзакций.
// Параметры:
//   - numberOfTx: количество последних транзакций для получения.
//
// Возвращает:
//   - []model.Transaction: массив транзакций.
//   - error: ошибку при выполнении запроса.
func (tr *TransactionRepository) GetLastTransactions(numberOfTx int) ([]model.Transaction, error) {
	query :=
		`SELECT id, from_wallet, to_wallet, amount, transfer_date
	FROM transactions
	ORDER BY transfer_date DESC 
	LIMIT $1;`

	rows, err := tr.db.Query(query, numberOfTx)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

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

// GetWallet возвращает информацию о кошельке по его ID.
// Параметры:
//   - walletId: идентификатор кошелька.
//
// Возвращает:
//   - *model.Wallet: структура кошелька с балансом и датой обновления.
//   - error: ошибку при выполнении запроса.
func (wr *WalletRepository) GetWallet(walletId uuid.UUID) (*model.Wallet, error) {
	db := wr.db

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query :=
		`SELECT
		w.id,
		w.balance,
		w.date_update
	FROM wallets w
	WHERE w.id = $1;`

	var wallet model.Wallet
	var balanceStr string

	err = tx.QueryRow(query, walletId).Scan(&wallet.Id, &balanceStr, &wallet.DateUpdate)
	if err != nil {
		return nil, err
	}

	balance := new(big.Float)
	if _, ok := balance.SetString(balanceStr); !ok {
		return nil, fmt.Errorf("failed to parse balance: %s", balanceStr)
	}
	wallet.Balance = *balance

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &wallet, nil
}
