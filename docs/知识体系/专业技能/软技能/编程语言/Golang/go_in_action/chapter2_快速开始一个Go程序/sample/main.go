package main

import (
	"log"
	"os"

	_ "github.com/goinaction/code/chapter2/sample/matchers"
	"github.com/goinaction/code/chapter2/sample/search"
)

// init 在main之前调用
func init() {
	// 日志输出到标准输出
	log.SetOutput(os.Stdout)
}

// main 整个程序的入口
func main() {
	search.Run("president")
}
