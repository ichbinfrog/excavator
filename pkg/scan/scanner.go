package scan

import (
	"sync"

	"github.com/ichbinfrog/excavator/pkg/model"
)

type Scanner interface {
	Scan(concurrent int)
	Type() string
}

func leakReader(leaksChan <-chan model.Leak, doneChan <-chan bool, task *sync.WaitGroup, res [][]model.Leak, idx int) {
	leaks := []model.Leak{}
	for {
		select {
		case leak := <-leaksChan:
			leaks = append(leaks, leak)
		case <-doneChan:
			res[idx] = leaks
			task.Done()
			break
		}
	}
}
