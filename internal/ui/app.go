package ui

import (
	"fmt"
	"os" // 用于获取当前工作目录以保存配置
	"strings"
	"time" // 用于 Tick

	"github.com/DoraleCitrus/gentr/internal/core"
	"github.com/DoraleCitrus/gentr/internal/model"
	"github.com/atotto/clipboard"                // 剪贴板库
	"github.com/charmbracelet/bubbles/textinput" // 输入框组件
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
	// 注释样式：深灰色 + 斜体
	annotationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true)
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

	// 用于在状态栏显示临时消息
	StatusMsg string

	// 输入框相关状态
	TextInput textinput.Model // 输入框组件
	InputMode bool            // 是否处于编辑模式
}

// InitialModel 初始化状态
func InitialModel(root *model.Node, limitReached bool) MainModel {
	// 初始化输入框
	ti := textinput.New()
	ti.Placeholder = "Type comment..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return MainModel{
		RootNode:     root,
		Cursor:       0,
		Quitting:     false,
		Width:        80,
		Height:       24,
		LimitWarning: limitReached, // 注入状态
		StatusMsg:    "",           // 初始化为空
		TextInput:    ti,           // 注入输入框
		InputMode:    false,        // 默认关闭
	}
}

// Init 是程序启动时执行的初始化方法
func (m MainModel) Init() tea.Cmd {
	// 启动时进入 AltScreen 全屏模式，解决拖动残留问题
	return tea.EnterAltScreen
}

// Update 处理用户输入并更新状态
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 定义 cmd 变量用于处理 bubbles 组件的命令
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// 监听窗口大小变化
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	// 区分 输入模式/导航模式
	if m.InputMode {
		// 输入模式
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				// 保存注释
				idx := 0
				node := m.getNodeAtCursor(m.RootNode.Children, &idx)
				if node != nil {
					node.Annotation = m.TextInput.Value()
					m.saveState() // 保存状态到文件
				}
				m.InputMode = false
				m.StatusMsg = "Comment saved!"
				return m, nil

			case "esc":
				// 取消编辑
				m.InputMode = false
				m.StatusMsg = "Cancelled."
				return m, nil
			}
		}
		// 让输入框组件处理具体的打字逻辑
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd

	} else {
		// 导航模式
		switch msg := msg.(type) {
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
					m.StatusMsg = "" // 移动光标时清除提示消息
				}

			// 向下移动光标
			case "down", "j":
				// 限制光标不能超过文件树的总行数
				// 我们需要计算一下当前可见的总节点数
				totalNodes := m.countVisibleNodes(m.RootNode.Children)
				if m.Cursor < totalNodes-1 {
					m.Cursor++
					m.StatusMsg = "" // 移动光标时清除提示消息
				}

			// 空格键折叠/展开
			case " ":
				idx := 0
				// 传入 idx 指针，在递归中寻找当前光标对应的节点
				// 如果发生状态改变，触发保存
				if m.toggleNode(m.RootNode.Children, &idx) {
					m.saveState() // 保存状态到文件
				}

			// 回车键隐藏/显示
			case "enter":
				idx := 0
				// 如果发生状态改变，触发保存
				if m.toggleHidden(m.RootNode.Children, &idx) {
					m.saveState() // 保存状态到文件
				}

			// 'c' 键复制功能
			case "c":
				output := m.generateTreeOutput()
				// 使用 markdown 代码块包裹，方便直接粘贴到文档
				finalText := fmt.Sprintf("```text\n%s```", output)

				err := clipboard.WriteAll(finalText)
				if err != nil {
					m.StatusMsg = "Error copying to clipboard!"
				} else {
					m.StatusMsg = "Copied to clipboard!"
				}
				// 返回一个空的 Tick 强制触发 View 刷新以显示 StatusMsg
				return m, tea.Tick(time.Millisecond, func(t time.Time) tea.Msg { return nil })

			// 按 'i' 进入编辑模式
			case "i":
				idx := 0
				node := m.getNodeAtCursor(m.RootNode.Children, &idx)
				if node != nil {
					m.InputMode = true
					// 把当前已有的注释填进去，方便修改
					m.TextInput.SetValue(node.Annotation)
					// 让光标闪烁
					return m, textinput.Blink
				}
			}
		}
	}
	return m, nil
}

// saveState 保存当前状态到 .gentr.json
func (m MainModel) saveState() {
	cwd, _ := os.Getwd()
	// 调用 core 包的 SaveConfig 方法
	// 频繁 IO 可能需要做防抖处理，
	// 但 TUI 操作频率较低，直接同步保存通常没有性能问题。
	_ = core.SaveConfig(cwd, m.RootNode)
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

	// 底部区域逻辑：根据模式切换显示内容
	bottomBar := ""

	if m.InputMode {
		// 1. 如果在输入模式，显示输入框
		bottomBar = fmt.Sprintf("\nAdding comment for selected file:\n%s\n(Enter to save, Esc to cancel)", m.TextInput.View())
	} else {
		// 2. 如果在导航模式，显示状态栏 + 帮助
		// 状态栏逻辑：优先显示 StatusMsg
		statusText := ""
		if m.StatusMsg != "" {
			statusText = m.StatusMsg // 显示 "Copied!" 等消息
		} else {
			// 没有系统消息时，显示当前文件路径
			idx := 0
			currentNode := m.getNodeAtCursor(m.RootNode.Children, &idx)
			if currentNode != nil {
				statusText = fmt.Sprintf("PATH: %s", currentNode.Path)
			} else {
				statusText = "(No selection)"
			}
		}

		// 只有当宽度足够时才进行截断操作，防止 Panic
		if m.Width > 5 && len(statusText) > m.Width-2 {
			statusText = statusText[:m.Width-5] + "..."
		}

		// 使用 statusBarStyle 渲染状态栏，并让它占满整行宽度
		statusBar := statusBarStyle.Width(m.Width).Render(statusText)

		// 提示文案
		help := "\n[Space] Toggle  [Enter] Hide/Show  [i] Comment  [c] Copy  [↑/↓] Move  [q] Quit"
		bottomBar = statusBar + help
	}

	result := warningBar + header + treeView + "\n" + bottomBar

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

		// 处理注释的显示逻辑
		annotationStr := ""
		if child.Annotation != "" {
			annotationStr = fmt.Sprintf("  # %s", child.Annotation)
		}

		// 增加对极小宽度的判断，防止 availableWidth < 0 导致 crash
		if availableWidth <= 1 {
			displayName = "" // 空间太小，直接不显示
			annotationStr = ""
		} else {
			// 计算总内容宽度 (名字 + 注释)
			totalWidth := lipgloss.Width(displayName + annotationStr)

			// 如果总宽度超过可用空间，需要截断
			if totalWidth > availableWidth {
				nameWidth := lipgloss.Width(displayName)

				// 情况1: 连名字都放不下
				if nameWidth >= availableWidth {
					annotationStr = "" // 不显示注释
					runes := []rune(displayName)
					// 截断名字
					truncateLen := availableWidth - 1
					if truncateLen > 0 && truncateLen < len(runes) {
						displayName = string(runes[:truncateLen]) + "…"
					}
				} else {
					// 情况2: 名字放得下，但注释放不下，截断注释
					// 剩余给注释的空间
					remainForAnno := availableWidth - nameWidth
					runesAnno := []rune(annotationStr)
					if remainForAnno > 1 && remainForAnno < len(runesAnno) {
						annotationStr = string(runesAnno[:remainForAnno-1]) + "…"
					} else {
						annotationStr = "" // 空间太小连省略号都放不下
					}
				}
			}
		}

		// 拼接字符串：光标指示器 + 缩进 + 连接符 + 文件名 + [注释]
		// 单独渲染注释部分
		line := fmt.Sprintf("%s%s%s%s%s%s",
			cursorIndicator,
			dimmedStyle.Render(prefix),
			dimmedStyle.Render(connector),
			icon,
			style.Render(displayName),
			annotationStyle.Render(annotationStr), // 渲染注释
		)
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

// generateTreeOutput 生成纯文本树，复制到剪贴板用
func (m MainModel) generateTreeOutput() string {
	var sb strings.Builder
	// 根目录不带前缀
	sb.WriteString(fmt.Sprintf("%s\n", m.RootNode.Name))

	// 递归生成子节点，过滤掉 Hidden 的
	// 不需要 index 指针也不需要 cursor 逻辑，只需要纯粹的遍历
	sb.WriteString(m.generateChildrenText(m.RootNode.Children, ""))
	return sb.String()
}

// 递归生成纯文本内容
func (m MainModel) generateChildrenText(children []*model.Node, prefix string) string {
	var sb strings.Builder

	// 1. 先过滤：找出所有没被隐藏的子节点
	// 必须先知道有哪些节点是“可见”的，才能正确计算谁是“最后一个”，保证线条不断裂
	var visibleChildren []*model.Node
	for _, child := range children {
		if !child.Hidden {
			visibleChildren = append(visibleChildren, child)
		}
	}

	// 2. 遍历可见节点
	for i, child := range visibleChildren {
		isLast := i == len(visibleChildren)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// 输出行：前缀 + 连接线 + 文件名 + [注释]
		line := fmt.Sprintf("%s%s%s", prefix, connector, child.Name)
		if child.Annotation != "" {
			// 导出时的注释格式，用空格对齐
			line += fmt.Sprintf("  # %s", child.Annotation)
		}
		sb.WriteString(line + "\n")

		// 递归处理子文件夹
		// 如果用户在界面上折叠了，导出时也不希望看到展开的细节，符合所见即所得
		if child.IsDir && len(child.Children) > 0 && !child.Collapsed {
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			sb.WriteString(m.generateChildrenText(child.Children, childPrefix))
		}
	}
	return sb.String()
}
