package model

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"time"
)

type TransferMoneyResponse struct {
	HttpStatus    int
	TransactionId uuid.UUID
}

type TransactionInfoResponse struct {
	TransactionId uuid.UUID `json:"transactionId"`
	From          uuid.UUID `json:"from"`
	To            uuid.UUID `json:"to"`
	Amount        BigFloat  `json:"amount"`
	TransferDate  time.Time `json:"transferDate"`
}

type WalletResponse struct {
	Id         uuid.UUID `json:"id"`
	Balance    BigFloat  `json:"balance"`
	DateUpdate time.Time `json:"date_update"`
}

type BigFloat big.Float

func (f *BigFloat) MarshalJSON() ([]byte, error) {
	s := (*big.Float)(f).Text('f', -1)
	return json.Marshal(s)
}

func (f *BigFloat) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	_, ok := (*big.Float)(f).SetString(s)
	if !ok {
		return fmt.Errorf("invalid big.Float string: %s", s)
	}
	return nil
}
