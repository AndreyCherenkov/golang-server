package model

import (
	"github.com/google/uuid"
)

type TransferMoneyRequest struct {
	From   uuid.UUID `json:"from"`
	To     uuid.UUID `json:"to"`
	Amount string    `json:"amount"`
}
