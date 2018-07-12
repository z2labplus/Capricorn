package main

import "fmt"

// notifier是一个定义了通知类型为的接口
type notifier interface {
	notify()
}

// user在程序里定义一个用户类型
type person struct {
	name  string
	email string
}

// notify 是使用指针接受者实现的方法
func (p *person) notify() {
	fmt.Printf("Sending User Email By Pointer Type Method.\n")
	fmt.Printf("Sending User Email %s <%s> \n", p.name, p.email)
}

// main 是程序的入口
func main() {
	// 创建一个user类型的值，并发送通知
	p := person{"Bill", "bill@email.com"}

	sendNotification(&p)
}

// sendNotification 接收一个实现了notifier接口的值，并发送通知
func sendNotification(n notifier) {
	n.notify()
}
