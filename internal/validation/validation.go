package validation

import (
	"errors"
	"strconv"
)

func ValidateAmount(amount string) error {
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return errors.New("invalid amount: must be a number")
	}
	if amountFloat <= 0 {
		return errors.New("amount must be greater than zero")
	}

	return nil
}
