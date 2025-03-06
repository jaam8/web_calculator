package orchestrator

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// ValidateExpression валидирует выражение
func ValidateExpression(expr string) error {
	// Удаляем пробелы
	expr = strings.ReplaceAll(expr, " ", "")
	if expr == "" {
		return errors.New("")
	}

	// Регулярка для проверки допустимых символов (цифры, операторы, скобки)
	validExpr := regexp.MustCompile(`^[0-9+\-*/().\s]+$`)
	if !validExpr.MatchString(expr) {
		return errors.New("")
	}

	// Проверка сбалансированности скобок
	if !isBracketsBalanced(expr) {
		return errors.New("")
	}

	// Проверка, что не начинается/заканчивается оператором
	if strings.HasPrefix(expr, "+") || strings.HasPrefix(expr, "-") || strings.HasPrefix(expr, "*") || strings.HasPrefix(expr, "/") ||
		strings.HasSuffix(expr, "+") || strings.HasSuffix(expr, "-") || strings.HasSuffix(expr, "*") || strings.HasSuffix(expr, "/") {
		return errors.New("")
	}

	// Проверка, что нет двух операторов подряд
	if match, _ := regexp.MatchString(`[+\-*/]{2,}`, expr); match {
		return errors.New("")
	}

	// Ищем выражения вида "/0" или "/0.0"
	divByZero := regexp.MustCompile(`/\s*0(\.0+)?\s*$`)
	if divByZero.MatchString(expr) {
		return errors.New("")
	}

	return nil
}

// isBracketsBalanced проверяет баланс круглых скобок
func isBracketsBalanced(expr string) bool {
	count := 0
	for _, ch := range expr {
		if ch == '(' {
			count++
		} else if ch == ')' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

var precedence = map[string]int{
	"+": 1, "-": 1,
	"*": 2, "/": 2,
}

// RPN преобразует выражение в обратную польскую нотацию
func RPN(expression string) ([]string, error) {
	var stack []string
	var output []string
	var num string
	if err := ValidateExpression(expression); err != nil {
		return nil, errors.New("")
	}
	for _, s := range expression {
		value := string(s)
		if unicode.IsDigit(s) || (value == "." && num != "") {
			num += value
		} else {
			if num != "" {
				output = append(output, num)
				num = ""
			}
			if _, ok := precedence[value]; ok {
				for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[value] {
					output = append(output, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				stack = append(stack, value)
			} else if value == "(" {
				stack = append(stack, value)
			} else if value == ")" {
				for len(stack) > 0 && stack[len(stack)-1] != "(" {
					output = append(output, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				if len(stack) == 0 {
					return nil, errors.New("")
				}
				stack = stack[:len(stack)-1]
			} else if !unicode.IsSpace(s) {
				return nil, errors.New("")
			}
		}
	}

	if num != "" {
		output = append(output, num)
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, errors.New("")
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}
