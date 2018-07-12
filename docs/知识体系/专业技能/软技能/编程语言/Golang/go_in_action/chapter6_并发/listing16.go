// 这个程序展示如何使用互斥锁
// 定义一段需要同步访问的代码临界区
// 资源的同步访问
package main

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	counter int
	wg      sync.WaitGroup

	// mutex 用来定义一段代码临界区
	mutex sync.Mutex
)

func main() {
	wg.Add(2)

	go incCounter(1)
	go incCounter(2)

	wg.Wait()

	fmt.Printf("Final Counter: %d \n", counter)
}

// incCounter 使用互斥锁来同步并保证安全访问
// 增加包里的counter变量的值

func incCounter(id int) {
	defer wg.Done()

	for count := 0; count < 2; count++ {
		// 同一时刻只允许一个goroutine进入这个临界区
		mutex.Lock()
		{
			value := counter
			runtime.Gosched()

			value++

			counter = value
		}
		// 释放锁，允许其他正在等待的goroutine进入临界区
		mutex.Unlock()
	}
}

// 同一时刻只有一 个 goroutine 可以进入临界区。
// 之后，直到调用 Unlock()函数之后，其他 goroutine 才能进入临 界区。
// 当强制将当前 goroutine 退出当前线程后，调度器会再次分配这个 goroutine 继续运 行。
// 当程序结束时，我们得到正确的值 4，竞争状态不再存在。
