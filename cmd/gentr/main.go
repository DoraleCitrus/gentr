package main

import (
	"fmt"
	"os"
)

func main() {
	// 简单的欢迎界面(测试环境)
	fmt.Println("Welcome to Gentr!")
	fmt.Println("---------------------")

	// 获取并打印当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	fmt.Printf("Scanning directory: %s\n", dir)
}
