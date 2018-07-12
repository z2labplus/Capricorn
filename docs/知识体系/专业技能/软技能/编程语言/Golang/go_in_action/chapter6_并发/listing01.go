package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	// 分配一个逻辑处理器给调度器使用
	// 调用了 runtime 包的 GOMAXPROCS 函数。
	// 这个函数允许程序 更改调度器可以使用的逻辑处理器的数量。
	// 如果不想在代码里做这个调用，也可以通过修改和这个函数名字一样的环境变量的值来更改逻辑处理器的数量。
	// 给这个函数传入1，是通知调度器只 能为该程序使用一个逻辑处理器。
	runtime.GOMAXPROCS(2)

	// wg用来等待程序完成
	// 计数加2，表示要等待两个goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("Start Goroutines")

	// 声明一个匿名函数，并创建一个goroutine
	go func() {
		// 在函数退出时，调用Done来通知main函数工作已经完成
		defer wg.Done()

		// 显示字母表3次
		for count := 0; count < 3; count++ {
			for char := 'a'; char < 'a'+26; char++ {
				fmt.Printf("%c ", char)
			}
		}
	}()

	// 声明一个匿名函数，并创建一个goroutine
	go func() {
		// 在函数退出时，调用Done来通知main函数工作已经完成
		// 关键字 defer 会修改函数调用时机，在正在执行的函数返回时才真正调用 defer 声明的函 数。
		// 对这里的示例程序来说，我们使用关键字 defer 保证，每个 goroutine 一旦完成其工作就调 用 Done 方法。
		defer wg.Done()
		// 基于调度器的内部算法，一个正运行的 goroutine 在工作结束前，可以被停止并重新调度。
		// 调度器这样做的目的是防止某个 goroutine 长时间占用逻辑处理器。
		// 当 goroutine 占用时间过长时， 调度器会停止当前正运行的 goroutine，并给其他可运行的 goroutine 运行的机会。

		// 显示字母表3次
		for count := 0; count < 3; count++ {
			for char := 'A'; char < 'A'+26; char++ {
				fmt.Printf("%c ", char)
			}
		}
	}()

	// 等待goroutine结束
	fmt.Println("Waiting To Finish")

	// WaitGroup 是一个计数信号量，可以用来记录并维护运行的 goroutine。
	// 如果 WaitGroup 的值大于 0，Wait 方法就会阻塞。
	wg.Wait()

	fmt.Println("\nTerminating Program")
}
