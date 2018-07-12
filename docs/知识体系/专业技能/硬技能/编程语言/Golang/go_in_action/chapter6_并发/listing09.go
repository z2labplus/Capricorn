package main

import (
	"fmt"
	"runtime"
	"sync"
)

// 这个示例程序展示如何在程序里造成竞争状态
// 实际上不希望出现这种情况

var (
	// counter是所有goroutine都要增加其值的变量
	counter int

	// wg 用来等待程序结束
	wg sync.WaitGroup
)

// main 是所有程序的入口
func main() {
	// 计数加2，表示要等待两个goroutine
	wg.Add(2)

	// 创建两个goroutine
	go incCounter(1)
	go incCounter(2)

	// 等待goroutine结束
	wg.Wait()
	fmt.Println("Final Counter:", counter)
}

// incCounter增加包里counter变量的值
func incCounter(id int) {
	// 在函数退出时调用Done来通知main函数工作已经完成
	defer wg.Done()

	for count := 0; count < 2; count++ {
		// 捕获count的值
		value := count

		// 当前的goroutine从线程退出，并放回到队列
		runtime.Gosched()

		// 增加value的值
		value++

		// 将该值保存回counter
		counter = value
	}
}

// 每个 goroutine 都会覆盖另一个 goroutine 的工作。这种覆盖发生在 goroutine 切换的时候。
// 每 个 goroutine 创造了一个 counter 变量的副本，之后就切换到另一个 goroutine。
// 当这个 goroutine 再次运行的时候，counter 变量的值已经改变了，
// 但是 goroutine 并没有更新自己的那个副本的 值，而是继续使用这个副本的值，用这个值递增，并存回 counter 变量，结果覆盖了另一个 goroutine 完成的工作。
