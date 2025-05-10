package helper

import (
	"github.com/jaam8/web_calculator/common-lib/errors"
	"unicode"
)

var precedence = map[string]int{
	"+": 1, "-": 1,
	"*": 2, "/": 2,
}

// ToRPN преобразует выражение в обратную польскую нотацию
func ToRPN(expression string) ([]string, error) {
	var stack []string
	var output []string
	var num string
	if err := ValidateExpression(expression); err != nil {
		return nil, errors.ErrInvalidExpression
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
					return nil, errors.ErrInvalidExpression
				}
				stack = stack[:len(stack)-1]
			} else if !unicode.IsSpace(s) {
				return nil, errors.ErrInvalidExpression
			}
		}
	}

	if num != "" {
		output = append(output, num)
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, errors.ErrInvalidExpression
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}
