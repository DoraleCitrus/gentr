package ui

import (
	"fmt"
	"strings"

	"github.com/DoraleCitrus/gentr/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 定义样式
var (
	// 选中的行：粉色,加粗
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	// 普通的行：白色
	normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	// 辅助符号：暗色
	dimmedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// MainModel 是 TUI 的状态容器
type MainModel struct {
	RootNode *model.Node // 之前的扫描结果
	Cursor   int         // 记录当前光标在第几行
	Quitting bool        // 用户是否选择退出
}

// InitialModel 初始化状态
func InitialModel(root *model.Node) MainModel {
	return MainModel{
		RootNode: root,
		Cursor:   0,
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

		// 向上移动光标
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		// 向下移动光标
		case "down", "j":
			// 限制光标不能超过文件树的总行数
			// 我们需要计算一下当前可见的总节点数 (目前暂未实现折叠，所以是所有节点)
			totalNodes := m.countVisibleNodes(m.RootNode.Children)
			if m.Cursor < totalNodes-1 {
				m.Cursor++
			}

		// 空格键折叠/展开
		case " ":
			idx := 0
			// 传入 idx 指针，在递归中寻找当前光标对应的节点
			m.toggleNode(m.RootNode.Children, &idx)
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
	// 根节点本身不需要前缀,它的子节点开始要有前缀
	// 传递一个整数指针 &index 进去，在递归函数里，index 的值会累加，从而实现全局计数
	index := 0
	s += m.renderChildren(m.RootNode.Children, "", &index)

	s += "\n[Space] Toggle folder  [↑/↓] Move  [q] Quit\n"
	return s
}

// renderChildren 遍历一组子节点并生成字符串,并接收一个 *int 类型的 index 指针
func (m MainModel) renderChildren(children []*model.Node, prefix string, index *int) string {
	var sb strings.Builder
	for i, child := range children {
		// 判断是否是列表中的最后一个，这决定了使用 └── 还是 ├──
		isLast := i == len(children)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// 判断当前行是否光标所在行
		cursorIndicator := "  " // 默认没有光标指示符
		style := normalStyle    // 默认样式

		if *index == m.Cursor {
			cursorIndicator = "> " // 光标指示符
			style = selectedStyle  // 应用选中样式
		}

		// 文件夹指示处理
		icon := ""
		if child.IsDir {
			if child.Collapsed {
				icon = "▶ " // 折叠状态
			} else {
				icon = "▼ " // 展开状态
			}
		} else {
			icon = "  " // 文件没有图标
		}

		// 拼接字符串：光标指示器 + 缩进 + 连接符 + 文件名，e.g. "> │   ├── main.go"
		// 用 Lip Gloss 的 style.Render() 来给文件名上色
		line := fmt.Sprintf("%s%s%s%s%s", cursorIndicator, dimmedStyle.Render(prefix), dimmedStyle.Render(connector), icon, style.Render(child.Name))
		sb.WriteString(line + "\n")

		// 处理完一行, 递增 index
		*index++

		// 如果是文件夹且有子节点，递归渲染其子节点
		if child.IsDir && len(child.Children) > 0 && !child.Collapsed {
			// 计算新的前缀
			// 如果当前节点是最后一个，那子节点的缩进就是空格 "    "
			// 如果当前节点不是最后一个，那子节点的缩进还需要竖线 "│   " 来连接下面的兄弟节点
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			sb.WriteString(m.renderChildren(child.Children, childPrefix, index))
		}
	}
	return sb.String()
}

// countNodes 计算当前可见节点的数量，用于防止光标越界
func (m MainModel) countVisibleNodes(children []*model.Node) int {
	count := 0
	for _, child := range children {
		count++
		// 只有没折叠的目录，才把它的子节点算进总数里
		if child.IsDir && !child.Collapsed {
			count += m.countVisibleNodes(child.Children)
		}
	}
	return count
}

// toggleNode 找到光标位置的节点并切换状态
// 和渲染一样，模拟遍历一遍，到 m.Cursor 就停下来执行切换操作
func (m MainModel) toggleNode(children []*model.Node, index *int) bool {
	for _, child := range children {
		if *index == m.Cursor {
			// 找到目标
			if child.IsDir {
				child.Collapsed = !child.Collapsed
			}
			return true // 告诉上层找到了，停止搜索
		}

		*index++

		if child.IsDir && !child.Collapsed {
			// 递归调用子节点，并接收返回值
			found := m.toggleNode(child.Children, index)
			if found {
				return true // 如果子节点里找到了也立刻停止，告诉上层
			}
		}
	}
	return false // 这一层没找到，继续找兄弟节点
}
