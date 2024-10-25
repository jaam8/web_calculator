package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	fmt.Println(Calc("1+1*"))
}

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
				return []string{}, errors.New("Неопознанный символ")
			}
		}
	}
	if num != "" {
		nums = append(nums, num)
	}
	return nums, nil
}

// структура описывающая приоритет оператора и функцию выполняющую действие оператора
type priorOper struct {
	priority  uint8
	operation func(float64, float64) float64
}

// функция возвращающая итоговый результат
func Calc(expression string) (float64, error) {
	ops := map[string]priorOper{
		"+": priorOper{1, func(a, b float64) float64 { return a + b }},
		"-": priorOper{1, func(a, b float64) float64 { return a - b }},
		"*": priorOper{2, func(a, b float64) float64 { return a * b }},
		"/": priorOper{2, func(a, b float64) float64 { return a / b }},
	}

	operators := "+-*/"
	var actions []string
	var nums []float64

	// производит действие над последними элементами
	applyActions := func() error {
		if len(nums) < 2 {
			return errors.New("Неправильное выражение")
		}

		b := nums[len(nums)-1]
		a := nums[len(nums)-2]
		op := actions[len(actions)-1]

		nums = nums[:len(nums)-2]
		actions = actions[:len(actions)-1]

		if op == "/" && b == 0 {
			return errors.New("Деление на ноль")
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
			last_action := ""
			if len(actions) > 0 {
				last_action = actions[len(actions)-1]
			} 

			for len(actions) > 0 &&
				strings.Contains(operators, string(last_action)) &&
				ops[last_action].priority >= ops[value].priority {

				if res := applyActions(); res != nil {
					return 0, res
				}
			}
			actions = append(actions, string(value))
		} else if value == "(" {
			actions = append(actions, string(value))
		} else if value == ")" {
			last_action := actions[len(actions)-1]
			for len(actions) > 0 && last_action != "(" {

				if res := applyActions(); res != nil {
					return 0, res
				}
				last_action = actions[len(actions)-1]
			}
			actions = actions[:len(actions)-1]
			// if len(actions) == 0 || last_action != "(" {
			// 	return 0, errors.New("Несоответствие скобок")
			// }
		} else {
			return 0, fmt.Errorf("Неизвестный символ: %v", value)
		}
		i += 1
	}

	for len(actions) > 0 {
		last_action := actions[len(actions)-1]
		if last_action == "(" {
			return 0, errors.New("Несоответствие скобок")
		}
		if res := applyActions(); res != nil {
			return 0, res
		}
	}

	if len(nums) != 1 {
		return 0, errors.New("Неправильное выражение")
	}
	return nums[0], nil
}
