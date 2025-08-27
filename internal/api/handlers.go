package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang-server/internal/model"
	"golang-server/internal/service"
	"golang-server/internal/storage/postgres"
	"golang-server/internal/validation"
	"net/http"
	"regexp"
	"strconv"
)

// walletBalanceRegex соответствует пути: /api/wallet/{uuid}/balance
var walletBalanceRegex = regexp.MustCompile(`^/api/wallet/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12})/balance$`)

// TransactionHandler обрабатывает запросы, связанные с транзакциями.
type TransactionHandler struct {
	transactionService service.TransactionService
}

// WalletHandler обрабатывает запросы, связанные с кошельками.
type WalletHandler struct {
	walletService service.WalletService
}

// NewTransactionHandler создаёт новый обработчик транзакций.
func NewTransactionHandler(db *postgres.PgDB) *TransactionHandler {
	return &TransactionHandler{transactionService: service.NewTransactionService(db)}
}

// NewWalletHandler создаёт новый обработчик кошельков.
func NewWalletHandler(db *postgres.PgDB) *WalletHandler {
	return &WalletHandler{walletService: service.NewWalletService(db)}
}

// ServeHTTP маршрутизирует запросы для TransactionHandler.
func (th *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/send":
		errorMiddleware(th.sendMoney)(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/transactions":
		errorMiddleware(th.getLastTransactions)(w, r)
	default:
		http.NotFound(w, r)
	}
}

// ServeHTTP маршрутизирует запросы для WalletHandler.
func (wh *WalletHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && walletBalanceRegex.MatchString(r.URL.Path):
		errorMiddleware(wh.getWalletInfo)(w, r)
	default:
		http.NotFound(w, r)
	}
}

// sendMoney обрабатывает перевод средств между кошельками.
func (th *TransactionHandler) sendMoney(w http.ResponseWriter, r *http.Request) error {
	var req model.TransferMoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return newHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to parse JSON: %v", err))
	}

	if err := validation.ValidateAmount(req.Amount); err != nil {
		return newHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := th.transactionService.SendMoney(req)
	if err != nil {
		return newHTTPError(resp.HttpStatus, err.Error())
	}

	writeJSON(w, resp.HttpStatus, resp)
	return nil
}

// getLastTransactions возвращает последние N транзакций.
func (th *TransactionHandler) getLastTransactions(w http.ResponseWriter, r *http.Request) error {
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		return newHTTPError(http.StatusBadRequest, "invalid query parameter 'count'")
	}

	transactions, err := th.transactionService.GetLastTransactions(count)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err.Error())
	}

	writeJSON(w, http.StatusOK, transactions)
	return nil
}

// getWalletInfo возвращает информацию о балансе кошелька.
func (wh *WalletHandler) getWalletInfo(w http.ResponseWriter, r *http.Request) error {
	matches := walletBalanceRegex.FindStringSubmatch(r.URL.Path)
	if len(matches) != 2 {
		return newHTTPError(http.StatusBadRequest, "invalid wallet path")
	}

	walletID, err := uuid.Parse(matches[1])
	if err != nil {
		return newHTTPError(http.StatusBadRequest, "invalid wallet ID format")
	}

	info, err := wh.walletService.GetWalletInfo(walletID)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err.Error())
	}

	writeJSON(w, http.StatusOK, info)
	return nil
}
