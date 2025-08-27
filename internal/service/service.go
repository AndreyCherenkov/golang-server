package service

import (
	"github.com/google/uuid"
	"golang-server/internal/model"
	"golang-server/internal/storage"
	"golang-server/internal/storage/postgres"
	"net/http"
)

// TransactionService обрабатывает операции с транзакциями.
type TransactionService struct {
	walletRepository      storage.WalletRepository
	transactionRepository storage.TransactionRepository
}

// WalletService обрабатывает операции с кошельками.
type WalletService struct {
	walletRepository storage.WalletRepository
}

// NewTransactionService создаёт новый TransactionService с подключением к базе данных.
func NewTransactionService(db *postgres.PgDB) TransactionService {
	return TransactionService{
		walletRepository:      *storage.NewWalletRepository(db),
		transactionRepository: *storage.NewTransactionRepository(db),
	}
}

// NewWalletService создаёт новый WalletService с подключением к базе данных.
func NewWalletService(db *postgres.PgDB) WalletService {
	return WalletService{walletRepository: *storage.NewWalletRepository(db)}
}

// SendMoney выполняет перевод средств между кошельками.
// Проверяет наличие баланса у отправителя и возвращает HTTP-статус и ID транзакции.
func (ts *TransactionService) SendMoney(data model.TransferMoneyRequest) (model.TransferMoneyResponse, error) {
	id, err := ts.transactionRepository.SendMoney(data)
	if err != nil {
		return model.TransferMoneyResponse{HttpStatus: http.StatusInternalServerError}, err
	}

	return model.TransferMoneyResponse{HttpStatus: http.StatusOK, TransactionId: *id}, nil
}

// GetLastTransactions возвращает последние numberOfTx транзакций.
func (ts *TransactionService) GetLastTransactions(numberOfTx int) ([]model.TransactionInfoResponse, error) {
	tx, err := ts.transactionRepository.GetLastTransactions(numberOfTx)
	if err != nil {
		return nil, err
	}

	response := make([]model.TransactionInfoResponse, 0, len(tx))
	for _, t := range tx {
		response = append(response, model.TransactionInfoResponse{
			TransactionId: t.Id,
			From:          t.From,
			To:            t.To,
			Amount:        model.BigFloat(t.Amount),
			TransferDate:  t.TransferDate,
		})
	}

	return response, nil
}

// GetWalletInfo возвращает информацию о кошельке по его UUID.
func (ws *WalletService) GetWalletInfo(id uuid.UUID) (*model.WalletResponse, error) {
	r, err := ws.walletRepository.GetWallet(id)
	if err != nil {
		return nil, err
	}

	response := model.WalletResponse{
		Id:         r.Id,
		Balance:    model.BigFloat(r.Balance),
		DateUpdate: r.DateUpdate,
	}
	return &response, nil
}
