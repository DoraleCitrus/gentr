package main

import (
	"fmt"
	"os"

	"github.com/DoraleCitrus/gentr/internal/core"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Scanning directory: %s ...\n", cwd)

	// 调用 Walk 函数构建目录树
	rootNode, err := core.Walk(cwd)
	if err != nil {
		fmt.Printf("Error when walking tree: %v\n", err)
		return
	}

	// 简单打印根节点和一级子节点的数量来验证
	fmt.Println("Scan complete.")
	fmt.Printf("Root Name: %s\n", rootNode.Name)
	fmt.Printf("Total files/folders at root level (filtered): %d\n", len(rootNode.Children))
	fmt.Println("----------------")

	// 打印出一级文件名，看看是不是过滤了 .git
	for _, child := range rootNode.Children {
		prefix := "(File)"
		if child.IsDir {
			prefix = "(Dir) "
		}
		fmt.Printf("%s %s\n", prefix, child.Name)
	}
}
