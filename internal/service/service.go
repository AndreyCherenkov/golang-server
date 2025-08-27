package service

import (
	"fmt"
	"github.com/google/uuid"
	"golang-server/internal/model"
	"golang-server/internal/storage"
	"golang-server/internal/storage/postgres"
	"math/big"
	"net/http"
	"strconv"
)

type TransactionService struct {
	walletRepository      storage.WalletRepository
	transactionRepository storage.TransactionRepository
}

type WalletService struct {
	walletRepository storage.WalletRepository
}

func NewTransactionService(db *postgres.PgDB) TransactionService {
	return TransactionService{
		walletRepository:      *storage.NewWalletRepository(db),
		transactionRepository: *storage.NewTransactionRepository(db),
	}
}

func NewWalletService(db *postgres.PgDB) WalletService {
	return WalletService{walletRepository: *storage.NewWalletRepository(db)}
}

func (s *TransactionService) SendMoney(data model.TransferMoneyRequest) (model.TransferMoneyResponse, error) {
	senderBalance, err := s.walletRepository.GetBalance(data.From) //todo падает, если json (body) пуст
	if err != nil {
		return model.TransferMoneyResponse{HttpStatus: http.StatusBadRequest}, err
	}

	amount, err := strconv.ParseFloat(data.Amount, 64)
	if err != nil {
		return model.TransferMoneyResponse{HttpStatus: http.StatusBadRequest}, err
	}
	if senderBalance.Cmp(big.NewFloat(amount)) < 0 {
		return model.TransferMoneyResponse{HttpStatus: http.StatusBadRequest},
			fmt.Errorf("insufficient funds: balance %v, required %v", senderBalance, data.Amount)
	}

	id, err := s.transactionRepository.SendMoney(data)
	if err != nil {
		return model.TransferMoneyResponse{HttpStatus: http.StatusInternalServerError}, err
	}

	return model.TransferMoneyResponse{HttpStatus: http.StatusOK, TransactionId: *id}, nil
}

func (ts *TransactionService) GetLastTransactions(numberOfTx int) ([]model.TransactionInfoResponse, error) {
	tx, err := ts.transactionRepository.GetLastTransactions(numberOfTx)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response := make([]model.TransactionInfoResponse, 0, len(tx))
	for _, t := range tx {
		response = append(response, model.TransactionInfoResponse{
			TransactionId: t.Id,
			From:          t.From,
			To:            t.To,
			TransferDate:  t.TransferDate,
		})
	}

	return response, nil
}

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
