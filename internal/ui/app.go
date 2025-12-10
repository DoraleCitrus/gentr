package ui

import (
	"fmt"

	"github.com/DoraleCitrus/gentr/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

// MainModel 是 TUI 的状态容器
type MainModel struct {
	RootNode *model.Node // 之前的扫描结果
	Quitting bool        // 用户是否选择退出
}

// InitialModel 初始化状态
func InitialModel(root *model.Node) MainModel {
	return MainModel{
		RootNode: root,
		Quitting: false,
	}
}

// Init 是程序启动时执行的初始化方法
func (m MainModel) Init() tea.Cmd {
	return nil
}

// Update 处理用户输入并更新状态
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// 监听键盘事件
	case tea.KeyMsg:
		switch msg.String() {
		// 按 q 或 Ctrl+C 退出程序
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View 渲染终端上的界面
func (m MainModel) View() string {
	if m.Quitting {
		return "Exiting Gentr...\n"
	}

	// 先显示简单信息用于测试 UI 是否工作
	s := fmt.Sprintf("Project: %s\n\n", m.RootNode.Name)

	// 简单的遍历显示(仅一层)
	for _, child := range m.RootNode.Children {
		cursor := "  " // 以后这里会变成光标 ">"
		s += fmt.Sprintf("%s%s\n", cursor, child.Name)
	}

	s += "\nPress 'q' to quit.\n"
	return s
}
