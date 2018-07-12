// 程序展示如何使用有缓冲的通道和固定数目的
// goroutine来处理一堆工作
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	numberGoroutines = 4  // 要使用的goroutine的数量
	taskLoad         = 10 // 要处理的工作的数量
)

var wg sync.WaitGroup

func init() {
	// 初始化随机数种子
	rand.Seed(time.Now().Unix())
}

func main() {
	// 创建一个有缓冲的通道来管理工作
	tasks := make(chan string, taskLoad)

	// 启动goroutine来处理工作
	wg.Add(numberGoroutines)
	for gr := 1; gr <= numberGoroutines; gr++ {
		go worker(tasks, gr)
	}

	// 增加一组要完成的工作
	for post := 1; post <= taskLoad; post++ {
		tasks <- fmt.Sprintf("Tasks: %d", post)
	}

	// 当所有工作都处理完时关闭通道
	// 以便所有goroutine退出
	close(tasks)

	// 等待所有工作完成
	wg.Wait()
}

// worker 作为goroutine启动来处理
// 从有缓冲区的通道传入的工作
func worker(tasks chan string, worker int) {
	defer wg.Done()

	for {
		task, ok := <-tasks
		if !ok {
			fmt.Printf("Worker: %d: Shutting Down\n", worker)
			return
		}

		// 显示开始工作了
		sleep := rand.Int63n(100)
		time.Sleep(time.Duration((sleep)) * time.Millisecond)

		// 显示完成了工作
		fmt.Printf("Worker: %d : COmpleted %s:\n", worker, task)
	}
}
