package ui

import (
	"fmt"
	"strings"

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

	// 标题
	s := fmt.Sprintf("Project: %s\n", m.RootNode.Name)

	// 递归渲染文件树
	// 根节点本身不需要前缀,它的子节点开始要有层级
	s += m.renderChildren(m.RootNode.Children, "")

	s += "\nPress 'q' to quit.\n"
	return s
}

// renderChildren 遍历一组子节点并生成字符串
func (m MainModel) renderChildren(children []*model.Node, prefix string) string {
	var sb strings.Builder
	for i, child := range children {
		// 判断是否是列表中的最后一个，这决定了使用 └── 还是 ├──
		isLast := i == len(children)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// 拼接当前行: 前缀 + 连接符 + 文件/文件夹名, e.g. "│   ├── file.txt"
		sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, child.Name))

		// 如果是文件夹且有子节点，递归渲染其子节点
		if child.IsDir && len(child.Children) > 0 {
			// 计算新的前缀
			// 如果当前节点是最后一个，那子节点的缩进就是空格 "    "
			// 如果当前节点不是最后一个，那子节点的缩进还需要竖线 "│   " 来连接下面的兄弟节点
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			sb.WriteString(m.renderChildren(child.Children, childPrefix))
		}
	}
	return sb.String()
}
