package validation

import (
	"errors"
	"strconv"
)

// ValidateAmount проверяет корректность переданной строки с суммой.
// Функция выполняет следующие проверки:
// 1. Преобразует строку в число с плавающей запятой.
// 2. Проверяет, что сумма больше нуля.
// Параметры:
//   - amount: строковое представление суммы, которую нужно проверить.
//
// Возвращает:
//   - error: ошибку в случае некорректного формата или если сумма меньше или равна нулю.
//     Возвращает nil, если сумма корректна.
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
