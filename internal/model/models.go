package model

import (
	"github.com/google/uuid"
	"math/big"
	"time"
)

type Wallet struct {
	Id         uuid.UUID `json:"id"`
	Balance    big.Float `json:"balance"`
	DateUpdate time.Time `json:"date_update"`
}

type Transaction struct {
	Id           uuid.UUID `json:"id"`
	From         uuid.UUID `json:"from"`
	To           uuid.UUID `json:"to"`
	Amount       big.Float `json:"amount"`
	TransferDate time.Time `json:"transfer_date"`
}
