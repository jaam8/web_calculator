package agent

import (
	"github.com/jaam8/web_calculator/internal/config"
	o "github.com/jaam8/web_calculator/internal/orchestrator"
	"log"
	"sync"
	"time"
)

var conf = config.Configs

func Run() {
	computingPower := conf.ComputingPower

	var wg sync.WaitGroup

	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		log.Println("AGENT Starting worker ", i)
		go func() {
			defer wg.Done() // Уменьшаем счётчик при завершении
			Work()
		}()
	}
	wg.Wait()
}

func Work() {
	sleepTime := time.Duration(conf.WaitTime) * time.Second
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
			ID:     task.ID,
			Result: result,
		}

		err = PostResult(Result)
		if err != nil {
			log.Println("Error posting result:", err)
		}
		log.Printf("Task ID: %d processed, result: %f\n", task.ID, result)
		time.Sleep(sleepTime)
	}
}
