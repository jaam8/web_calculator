package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	o "github.com/jaam8/web_calculator/orchestrator"
	"io"
	"net/http"
	"strings"
	"time"
)

// DoTask вычисляет задачу
func DoTask(task o.Task) float64 {
	switch task.Operation {
	case "+":
		time.Sleep(task.OperationTime)
		return task.Arg1 + task.Arg2
	case "-":
		time.Sleep(task.OperationTime)
		return task.Arg1 - task.Arg2
	case "*":
		time.Sleep(task.OperationTime)
		return task.Arg1 * task.Arg2
	case "/":
		time.Sleep(task.OperationTime)
		if task.Arg2 == 0 {
			return 0
		}
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}

// GetTask делает запрос к оркестратору и возвращает задачу
func GetTask() (o.Task, int, error) {
	url := fmt.Sprintf("%s:%s/internal/task", conf.RequestURL, conf.Port)
	resp, err := http.Get(url)
	if err != nil {
		return o.Task{TaskID: 0}, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return o.Task{TaskID: 0}, resp.StatusCode, err
	}
	if resp.StatusCode != http.StatusOK {
		return o.Task{TaskID: 0}, resp.StatusCode, errors.New("task not found")
	}
	var task o.Task
	err = json.Unmarshal(body, &task)
	if err != nil {
		return o.Task{TaskID: 0}, resp.StatusCode, err
	}
	return task, resp.StatusCode, nil
}

// PostResult отправляет результат вычисления оркестратору
func PostResult(result o.Result) error {
	resultData, err := json.Marshal(result)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s:%s/internal/task", conf.RequestURL, conf.Port)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(resultData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		status := fmt.Sprintf("status %d\n", resp.StatusCode)
		path := fmt.Sprintf("path %s\n", resp.Request.URL)
		response := fmt.Sprintf("response %s\n", string(body))
		return fmt.Errorf("failed to post result: %s%s%s", status, path, response)
	}
	return nil
}
