package agent

import (
	"github.com/jaam8/web_calculator/internal/config"
	o "github.com/jaam8/web_calculator/internal/orchestrator"
	"log"
	"sync"
	"time"
)

var conf = config.Configs

// Run запускает агентов, является точкой входа для агентов
func Run() {
	computingPower := conf.ComputingPower

	var wg sync.WaitGroup

	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		log.Println("AGENT Starting worker ", i)
		go func() {
			defer wg.Done()
			Work()
		}()
	}
	wg.Wait()
}

// Work делает постоянные запросы к оркестратору
func Work() {
	sleepTime := time.Duration(conf.WaitTime) * time.Millisecond
	for {
		task, status, err := GetTask()
		for status != 200 {
			if err != nil && status == 404 {
				log.Println("Error getting task:", err)
				time.Sleep(sleepTime)
			}
			if err != nil && status == 500 {
				log.Println("Server error:", err)
				time.Sleep(sleepTime)
			}
			task, status, err = GetTask()
		}
		result := DoTask(task)

		Result := o.Result{
			ExpressionID: task.ExpressionID,
			TaskID:       task.TaskID,
			Result:       result,
		}

		err = PostResult(Result)
		if err != nil {
			log.Println("Error posting result:", err)
		}
		log.Printf("Task TaskID: %d processed, result: %f\n", task.TaskID, result)
		time.Sleep(sleepTime)
	}
}
