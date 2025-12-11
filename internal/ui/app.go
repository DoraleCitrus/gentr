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
	// 隐藏的行：灰色 + 删除线
	hiddenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Strikethrough(true)
	// 选中且隐藏的行：暗粉色 + 粗体 + 删除线
	selectedHiddenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("175")).Bold(true).Strikethrough(true)
	// 状态栏样式：背景白色文字紫色
	statusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("57")).Padding(0, 1)
	// 警告栏样式：黄色背景，黑色文字
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFFF00")).Padding(0, 1).Bold(true)
)

// MainModel 是 TUI 的状态容器
type MainModel struct {
	RootNode *model.Node // 之前的扫描结果
	Cursor   int         // 记录当前光标在第几行
	Quitting bool        // 用户是否选择退出

	// 终端的宽度和高度，用于计算截断
	Width  int
	Height int

	LimitWarning bool // 警告标记
}

// InitialModel 初始化状态
func InitialModel(root *model.Node, limitReached bool) MainModel {
	return MainModel{
		RootNode:     root,
		Cursor:       0,
		Quitting:     false,
		Width:        80,
		Height:       24,
		LimitWarning: limitReached, // 注入状态
	}
}

// Init 是程序启动时执行的初始化方法
func (m MainModel) Init() tea.Cmd {
	// 启动时进入 AltScreen 全屏模式，解决拖动残留问题
	return tea.EnterAltScreen
}

// Update 处理用户输入并更新状态
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// 监听窗口大小变化
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	// 监听键盘事件
	case tea.KeyMsg:
		switch msg.String() {
		// 按 q 或 Ctrl+C 退出程序
		case "q", "ctrl+c":
			m.Quitting = true
			// 退出时关闭 AltScreen，恢复终端原样
			return m, tea.ExitAltScreen

		// 向上移动光标
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		// 向下移动光标
		case "down", "j":
			// 限制光标不能超过文件树的总行数
			// 我们需要计算一下当前可见的总节点数
			totalNodes := m.countVisibleNodes(m.RootNode.Children)
			if m.Cursor < totalNodes-1 {
				m.Cursor++
			}

		// 空格键折叠/展开
		case " ":
			idx := 0
			// 传入 idx 指针，在递归中寻找当前光标对应的节点
			m.toggleNode(m.RootNode.Children, &idx)

		// 回车键隐藏/显示
		case "enter":
			idx := 0
			m.toggleHidden(m.RootNode.Children, &idx)
		}
	}
	return m, nil
}

// View 渲染终端上的界面
func (m MainModel) View() string {
	if m.Quitting {
		return "" // 退出时返回空字符串，避免屏幕闪烁
	}

	// 警告条逻辑
	warningBar := ""
	if m.LimitWarning {
		msg := "[!] Safety Limit Reached: Only showing first 5000 files / 10 levels deep."
		warningBar = warningStyle.Width(m.Width).Render(msg) + "\n"
	}

	// 标题
	header := fmt.Sprintf("Project: %s\n", m.RootNode.Name)

	// 递归渲染文件树
	// 根节点本身不需要前缀,它的子节点开始要有前缀
	// 传递一个整数指针 &index 进去，在递归函数里，index 的值会累加，从而实现全局计数
	index := 0
	// forceHidden 参数，初始为 false
	treeView := m.renderChildren(m.RootNode.Children, "", &index, false)

	// 状态栏
	idx := 0
	currentNode := m.getNodeAtCursor(m.RootNode.Children, &idx)
	statusText := "(No selection)"
	if currentNode != nil {
		statusText = fmt.Sprintf("PATH: %s", currentNode.Path)
	}

	// 只有当宽度足够时才进行截断操作，防止 Panic
	if m.Width > 5 && len(statusText) > m.Width-2 {
		statusText = statusText[:m.Width-5] + "..."
	}

	// 使用 statusBarStyle 渲染状态栏，并让它占满整行宽度
	statusBar := statusBarStyle.Width(m.Width).Render(statusText)

	// 提示文案
	help := "\n[Space] Toggle folder  [Enter] Hide/Show  [↑/↓] Move  [q] Quit"

	result := warningBar + header + treeView + "\n" + statusBar + help

	// 补齐空行，消除终端伪影
	lines := strings.Count(result, "\n") + 1
	if lines < m.Height {
		result += strings.Repeat("\n", m.Height-lines)
	}
	return result
}

// renderChildren 遍历一组子节点并生成字符串,并接收一个 *int 类型的 index 指针
// forceHidden bool 参数，用于处理父级隐藏时的级联效果
func (m MainModel) renderChildren(children []*model.Node, prefix string, index *int, forceHidden bool) string {
	var sb strings.Builder
	for i, child := range children {
		// 判断是否是列表中的最后一个，这决定了使用 └── 还是 ├──
		isLast := i == len(children)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// 判断当前节点是否应该显示为隐藏状态
		// 如果父节点强制隐藏(forceHidden) 或者 自身被标记隐藏(child.Hidden)
		isNodeHidden := forceHidden || child.Hidden

		// 判断当前行是否光标所在行
		cursorIndicator := "  " // 默认没有光标指示符
		style := normalStyle    // 默认样式

		// 样式逻辑：处理 普通/选中/隐藏/选中且隐藏 四种状态
		if isNodeHidden {
			style = hiddenStyle // 默认隐藏样式
		}

		if *index == m.Cursor {
			cursorIndicator = "> " // 光标指示符
			style = selectedStyle  // 默认选中样式
			// 选中且隐藏状态
			if isNodeHidden {
				style = selectedHiddenStyle
			}
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

		// 计算前缀的总可见长度 (Prefix + Connector + Icon + Cursor)
		prefixWidth := lipgloss.Width(cursorIndicator + prefix + connector + icon)

		// 计算留给文件名的剩余空间
		// 预留 1 个字符防止边缘溢出
		availableWidth := m.Width - prefixWidth - 1

		displayName := child.Name

		// 增加对极小宽度的判断，防止 availableWidth < 0 导致 crash
		if availableWidth <= 1 {
			displayName = "" // 空间太小，直接不显示文件名
		} else if lipgloss.Width(displayName) > availableWidth {
			// 核心截断逻辑
			runes := []rune(displayName)
			truncateLen := availableWidth - 1

			// 再次检查 truncateLen 是否合法
			if truncateLen > 0 && truncateLen < len(runes) {
				displayName = string(runes[:truncateLen]) + "…"
			}
		}

		// 拼接字符串：光标指示器 + 缩进 + 连接符 + 文件名
		line := fmt.Sprintf("%s%s%s%s%s", cursorIndicator, dimmedStyle.Render(prefix), dimmedStyle.Render(connector), icon, style.Render(displayName))
		sb.WriteString(line + "\n")

		// 处理完一行, 递增 index
		*index++

		// 如果是文件夹且有子节点，递归渲染其子节点
		if child.IsDir && len(child.Children) > 0 && !child.Collapsed {
			// 计算新的前缀
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			// 递归传递 isNodeHidden，实现级联隐藏
			sb.WriteString(m.renderChildren(child.Children, childPrefix, index, isNodeHidden))
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
func (m MainModel) toggleNode(children []*model.Node, index *int) bool {
	for _, child := range children {
		if *index == m.Cursor {
			if child.IsDir {
				child.Collapsed = !child.Collapsed
			}
			return true
		}
		*index++
		if child.IsDir && !child.Collapsed {
			if m.toggleNode(child.Children, index) {
				return true
			}
		}
	}
	return false
}

// toggleHidden 找到光标位置的节点并切换 Hidden 状态
func (m MainModel) toggleHidden(children []*model.Node, index *int) bool {
	for _, child := range children {
		if *index == m.Cursor {
			child.Hidden = !child.Hidden
			return true
		}
		*index++
		if child.IsDir && !child.Collapsed {
			if m.toggleHidden(child.Children, index) {
				return true
			}
		}
	}
	return false
}

// getNodeAtCursor 获取当前光标指向的节点对象，用于状态栏显示路径
func (m MainModel) getNodeAtCursor(children []*model.Node, index *int) *model.Node {
	for _, child := range children {
		if *index == m.Cursor {
			return child
		}
		*index++
		if child.IsDir && !child.Collapsed {
			found := m.getNodeAtCursor(child.Children, index)
			if found != nil {
				return found
			}
		}
	}
	return nil
}
