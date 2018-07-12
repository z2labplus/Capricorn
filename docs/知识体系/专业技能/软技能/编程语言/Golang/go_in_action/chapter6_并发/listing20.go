// 程序展示如何使用无缓冲的通道来模拟
// 2个goroutine间的网球比赛

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	court := make(chan int)

	wg.Add(2)

	// 启动两个选手
	go player("Nadal", court)
	go player("Djokovic", court)

	// 发球
	court <- 1

	wg.Wait()
}

// player 模拟一个选手在打网球
func player(name string, court chan int) {
	defer wg.Done()

	for {
		// 等待球被击打过来
		ball, ok := <-court
		if !ok {
			fmt.Printf("Player %s Won\n", name)
			return
		}

		// 选随机数，然后用这个数来判断我们是否丢球
		n := rand.Intn(100)

		if n%13 == 0 {
			fmt.Printf("Player %s Missed\n", name)

			close(court)
			return
		}

		fmt.Printf("Player %s Hit %d\n", name, ball)
		ball++

		court <- ball
	}
}
