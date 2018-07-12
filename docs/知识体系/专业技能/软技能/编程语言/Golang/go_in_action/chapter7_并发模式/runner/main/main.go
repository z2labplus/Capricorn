// 程序演示如何使用通道来监视
// 程序运行的时间，以在程序运行时间过长
// 时如何终止程序
package main

import (
	"github.com/goinaction/code/chapter7/patterns/runner"
	"log"
	"os"
	"time"
)

const timeout = 3 * time.Second

func main() {
	log.Println("Starting work.")

	r := runner.New(timeout)

	r.Add(createTask(), createTask(), createTask())

	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeout:
			log.Println("Terminating due to timeout")
			os.Exit(1)
		case runner.ErrInterrupt:
			log.Println("Terminating due to interrupt")
			os.Exit(2)
		}
	}

	log.Println("Process ended.")
}

func createTask() func(int2 int) {
	return func(id int) {
		log.Printf("Processor")
		time.Sleep(time.Duration(id) * time.Second)
	}
}
