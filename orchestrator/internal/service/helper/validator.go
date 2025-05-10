package helper

import (
	"github.com/jaam8/web_calculator/common-lib/errors"
	"regexp"
	"strings"
)

// ValidateExpression валидирует выражение
func ValidateExpression(expr string) error {
	// Удаляем пробелы
	expr = strings.ReplaceAll(expr, " ", "")
	if expr == "" {
		return errors.ErrInvalidExpression
	}

	// Регулярка для проверки допустимых символов (цифры, операторы, скобки)
	validExpr := regexp.MustCompile(`^[0-9+\-*/().\s]+$`)
	if !validExpr.MatchString(expr) {
		return errors.ErrInvalidExpression
	}

	// Проверка сбалансированности скобок
	if !isBracketsBalanced(expr) {
		return errors.ErrInvalidExpression
	}

	// Проверка, что не начинается/заканчивается оператором
	if strings.HasPrefix(expr, "+") || strings.HasPrefix(expr, "-") || strings.HasPrefix(expr, "*") || strings.HasPrefix(expr, "/") ||
		strings.HasSuffix(expr, "+") || strings.HasSuffix(expr, "-") || strings.HasSuffix(expr, "*") || strings.HasSuffix(expr, "/") {
		return errors.ErrInvalidExpression
	}

	// Проверка, что нет двух операторов подряд
	if match, _ := regexp.MatchString(`[+\-*/]{2,}`, expr); match {
		return errors.ErrInvalidExpression
	}

	if match, _ := regexp.MatchString(`\d*\.\d*\.\d*`, expr); match {
		return errors.ErrInvalidExpression
	}

	// Ищем выражения вида "/0" или "/0.0"
	divByZero := regexp.MustCompile(`/\s*0(\.0+)?\s*$`)
	if divByZero.MatchString(expr) {
		return errors.ErrDivideByZero
	}

	return nil
}

// isBracketsBalanced проверяет баланс круглых скобок
func isBracketsBalanced(expr string) bool {
	count := 0
	for _, ch := range expr {
		switch ch {
		case '(':
			count++
		case ')':
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}
