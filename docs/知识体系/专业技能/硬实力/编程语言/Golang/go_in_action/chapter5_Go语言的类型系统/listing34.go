package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// init 函数在main函数之前调用
func init() {
	if len(os.Args) != 2 {
		fmt.Println("Usage go run listing34.go <url>")
		os.Exit(-1)
	}
}

// main函数是程序的入库
func main() {
	var b bytes.Buffer
	// 将字符串写入Buffer
	b.Write([]byte("starting"))
	io.Copy(os.Stdout, &b)

	// 从Web服务器得到响应
	r, err := http.Get(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	// 从Body复制到stdout
	io.Copy(os.Stdout, r.Body)
	if err := r.Body.Close(); err != nil {
		fmt.Println(err)
	}
	log.Println("ending")
}
