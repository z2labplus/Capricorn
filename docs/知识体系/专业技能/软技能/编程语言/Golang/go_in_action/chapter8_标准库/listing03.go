package main

import "log"

func init() {
	log.SetPrefix("Trace: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}

func main() {
	// Println写到标准日志记录器
	log.Println("message")

	// Fatalln 在调用 Println()之后会接着调用 os.Exit(1)
	log.Fatalln("fatal message")

	// Panicln 在调用 Println()之后会接着调用 panic()
	log.Panicln("panic message")
}
