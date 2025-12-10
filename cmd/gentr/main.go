package main

import (
	"fmt"
	"os"

	"github.com/DoraleCitrus/gentr/internal/core"
	"github.com/DoraleCitrus/gentr/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 扫描文件
	cwd, _ := os.Getwd()
	rootNode, err := core.Walk(cwd)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	// 初始化 UI 模型
	initialModel := ui.InitialModel(rootNode)

	// 创建 Bubble Tea 程序并运行
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
