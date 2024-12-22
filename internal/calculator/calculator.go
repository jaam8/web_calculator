package calculator

import (
	"strconv"
	"strings"
	"unicode"
)

// конвертация expression в отдельные элементы для удобства
func convElem(expression string) ([]string, error) {
	operators := "+-*/()"
	var nums []string
	var num string

	for _, s := range expression {
		value := string(s)
		if unicode.IsDigit(s) || (value == "." && num != "") || (value == "-" && num == "") {
			num += string(value)
		} else {
			if num != "" {
				nums = append(nums, num)
				num = ""
			}
			if strings.Contains(operators, string(value)) {
				nums = append(nums, string(value))
			} else if unicode.IsDigit(s) {
				return []string{}, ErrInvalidExpression
			} else if !unicode.IsDigit(s) {
				return []string{}, ErrInvalidExpression
			}
		}
	}
	if num != "" {
		nums = append(nums, num)
	}
	return nums, nil
}

// структура описывающая приоритет оператора и функцию выполняющую действие оператора
type priorityOperation struct {
	priority  uint8
	operation func(float64, float64) float64
}

// Calculate функция возвращающая итоговый результат
func Calculate(expression string) (float64, error) {
	ops := map[string]priorityOperation{
		"+": priorityOperation{1, func(a, b float64) float64 { return a + b }},
		"-": priorityOperation{1, func(a, b float64) float64 { return a - b }},
		"*": priorityOperation{2, func(a, b float64) float64 { return a * b }},
		"/": priorityOperation{2, func(a, b float64) float64 { return a / b }},
	}

	operators := "+-*/"
	var actions []string
	var nums []float64

	// производит действие над последними элементами
	applyActions := func() error {
		if len(nums) < 2 {
			return ErrInvalidExpression
		}

		b := nums[len(nums)-1]
		a := nums[len(nums)-2]
		op := actions[len(actions)-1]

		nums = nums[:len(nums)-2]
		actions = actions[:len(actions)-1]

		if op == "/" && b == 0 {
			return ErrDivisionByZero
		}

		nums = append(nums, ops[op].operation(a, b))
		return nil
	}

	values, err := convElem(expression)

	if err != nil {
		return 0, err
	}

	i := 0
	for i < len(values) {
		value := values[i]

		if num, err := strconv.ParseFloat(value, 64); err == nil {
			nums = append(nums, num)
		} else if strings.Contains(operators, string(value)) {
			lastAction := ""
			if len(actions) > 0 {
				lastAction = actions[len(actions)-1]
			}

			for len(actions) > 0 &&
				strings.Contains(operators, string(lastAction)) &&
				ops[lastAction].priority >= ops[value].priority {

				if res := applyActions(); res != nil {
					return 0, res
				}
			}
			actions = append(actions, string(value))
		} else if value == "(" {
			actions = append(actions, string(value))
		} else if value == ")" {
			lastAction := actions[len(actions)-1]
			for len(actions) > 0 && lastAction != "(" {

				if res := applyActions(); res != nil {
					return 0, res
				}
				lastAction = actions[len(actions)-1]
			}
			actions = actions[:len(actions)-1]
		} else {
			return 0, ErrInvalidExpression
		}
		i += 1
	}

	for len(actions) > 0 {
		lastAction := actions[len(actions)-1]
		if lastAction == "(" {
			return 0, ErrInvalidExpression
		}
		if res := applyActions(); res != nil {
			return 0, res
		}
	}

	if len(nums) != 1 {
		return 0, ErrInvalidExpression
	}
	return nums[0], nil
}
