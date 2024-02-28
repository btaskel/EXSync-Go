package lan

import (
	"fmt"
	"sync"
)

type Workdist struct {
	Host string
}

const (
	taskload = 255
	tasknum  = 255
)

func Task(ip string) (onlineHosts []string) {
	onlineHostsChan := make(chan string, taskload)
	tasks := make(chan Workdist, taskload)
	var wg sync.WaitGroup
	wg.Add(tasknum)
	//创建chan消费者worker
	for gr := 1; gr <= tasknum; gr++ {
		go worker(&wg, tasks, onlineHostsChan)
	}

	//创建chan生产者
	for i := 1; i < 256; i++ {
		host := fmt.Sprintf("%s.%d", ip, i)
		fmt.Println(host)
		task := Workdist{
			Host: host,
		}
		tasks <- task
	}
	close(tasks)
	wg.Wait()
	close(onlineHostsChan)

	for onlineHost := range onlineHostsChan {
		onlineHosts = append(onlineHosts, onlineHost)
	}
	return
}

func worker(wg *sync.WaitGroup, tasks chan Workdist, onlineHostsChan chan string) {
	defer wg.Done()
	task, ok := <-tasks
	if !ok {
		return
	}
	ping(task.Host, onlineHostsChan)
}
