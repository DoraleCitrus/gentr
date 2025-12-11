package ui

import (
	"fmt"
	"os" // 用于获取当前工作目录以保存配置
	"runtime"
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
	// 搜索匹配的高亮样式 (黄色加粗)
	searchMatchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)

	// Git 样式
	// 黄色表示修改
	gitModifiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Bold(true)
	// 绿色表示新增
	gitAddedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")).Bold(true)
	// Git模式下的状态栏 (橙色背景)
	gitStatusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#D08770")).Padding(0, 1).Bold(true)

	// 更新横幅样式 (蓝色背景，白色文字，加粗)
	updateBannerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#005faff")).Padding(0, 1).Bold(true)
)

// SVG 主题系统定义
type Theme struct {
	Name         string
	BgColor      string // 背景色
	TextColor    string // 普通文件名颜色
	TreeColor    string // 树形线条颜色
	FolderColor  string // 文件夹图标颜色
	CommentColor string // 注释颜色
	GitModColor  string // [M] 颜色
	GitAddColor  string // [+] 颜色
}

// 预设两套风格主题
var (
	// Dark: 基于 VSCode Dark
	DarkTheme = Theme{
		Name:         "Dark",
		BgColor:      "#282a36",
		TextColor:    "#f8f8f2",
		TreeColor:    "#6272a4",
		FolderColor:  "#8be9fd",
		CommentColor: "#6272a4",
		GitModColor:  "#f1fa8c", // Yellow
		GitAddColor:  "#50fa7b", // Green
	}
	// Light: 基于 GitHub Light
	LightTheme = Theme{
		Name:         "Light",
		BgColor:      "#ffffff",
		TextColor:    "#24292e",
		TreeColor:    "#d1d5da", // Light grey for tree lines
		FolderColor:  "#0366d6", // Blue
		CommentColor: "#6a737d", // Grey
		GitModColor:  "#b08800", // Dark Yellow
		GitAddColor:  "#22863a", // Green
	}
)

// 定义防抖消息，携带版本号
type SaveMsg struct {
	Tag int
}

// 检查更新的消息
type CheckUpdateMsg struct {
	Available     bool
	LatestVersion string
}

// MainModel 是 TUI 的状态容器
type MainModel struct {
	RootNode     *model.Node // 之前的扫描结果
	Cursor       int         // 记录当前光标在第几行
	ScrollOffset int         // 滚动偏移量
	Quitting     bool        // 用户是否选择退出

	// 终端的宽度和高度，用于计算截断
	Width  int
	Height int

	LimitWarning bool // 警告标记

	// 用于在状态栏显示临时消息
	StatusMsg string

	// 输入框相关状态
	TextInput textinput.Model // 输入框组件
	InputMode bool            // 是否处于编辑模式

	// 搜索相关状态
	SearchInput textinput.Model
	SearchMode  bool

	// 防抖计数器 (版本号)
	SaveTag int

	// Git 模式开关
	GitMode bool

	// 版本相关字段
	CurrentVersion  string
	UpdateAvailable bool
	LatestVersion   string
}

// InitialModel 初始化状态
func InitialModel(root *model.Node, limitReached bool, currentVersion string) MainModel {
	// 初始化输入框
	ti := textinput.New()
	ti.Placeholder = "Type comment..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	// 初始化搜索输入框
	si := textinput.New()
	si.Placeholder = "Search files..."
	si.Prompt = "/ "
	si.CharLimit = 50
	si.Width = 50

	return MainModel{
		RootNode:       root,
		Cursor:         0,
		ScrollOffset:   0,
		Quitting:       false,
		Width:          80,
		Height:         24,
		LimitWarning:   limitReached,   // 注入状态
		StatusMsg:      "",             // 初始化为空
		TextInput:      ti,             // 注入输入框
		InputMode:      false,          // 默认关闭
		SearchInput:    si,             // 注入搜索框
		SearchMode:     false,          // 默认关闭
		SaveTag:        0,              // 防抖计数器初始化
		GitMode:        false,          // 默认关闭 Git 模式
		CurrentVersion: currentVersion, // 保存当前版本
	}
}

// Init 是程序启动时执行的初始化方法
func (m MainModel) Init() tea.Cmd {
	// 触发异步检查更新
	return m.checkUpdateCmd
}

// 检查更新的 Cmd
func (m MainModel) checkUpdateCmd() tea.Msg {
	available, latest, err := core.CheckForUpdates(m.CurrentVersion)
	if err != nil {
		// 网络错误等，静默失败，不打扰用户
		return CheckUpdateMsg{Available: false}
	}
	return CheckUpdateMsg{
		Available:     available,
		LatestVersion: latest,
	}
}

// viewportHeight 计算用于显示文件树的视口高度
func (m MainModel) viewportHeight() int {
	headerHeight := 1 // "Project: ..."
	if m.UpdateAvailable {
		headerHeight++
	}
	if m.LimitWarning {
		headerHeight++
	}

	footerHeight := 3 // Status bar + Help (approx)
	if m.InputMode || m.SearchMode {
		footerHeight = 4 // Input box + hint (approx)
	}

	h := m.Height - headerHeight - footerHeight
	if h < 1 {
		return 1
	}
	return h
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
		// 窗口大小改变时，确保光标可见
		vpHeight := m.viewportHeight()
		if m.Cursor < m.ScrollOffset {
			m.ScrollOffset = m.Cursor
		} else if m.Cursor >= m.ScrollOffset+vpHeight {
			m.ScrollOffset = m.Cursor - vpHeight + 1
		}

	// 处理防抖保存消息
	case SaveMsg:
		// 当消息里的 Tag 等于当前的 SaveTag 时，说明是最新的操作，执行保存
		if msg.Tag == m.SaveTag {
			m.saveStateImmediate()
		}
		return m, nil

	// 处理更新检查结果
	case CheckUpdateMsg:
		if msg.Available {
			m.UpdateAvailable = true
			m.LatestVersion = msg.LatestVersion
		}
		return m, nil
	}

	// 搜索模式优先处理
	if m.SearchMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "esc":
				// 退出搜索输入
				// Esc 清空并退出，Enter 仅退出输入焦点但保留过滤结果
				if msg.String() == "esc" {
					m.SearchInput.SetValue("") // 清空搜索
				}
				m.SearchMode = false
				m.Cursor = 0 // 退出搜索时重置光标防止越界
				m.ScrollOffset = 0
				return m, nil
			}
		}
		// 更新搜索输入框
		var siCmd tea.Cmd
		m.SearchInput, siCmd = m.SearchInput.Update(msg)

		// 每次按键后，搜索词变了，树的结构就变了，光标必须重置，防止越界
		m.Cursor = 0
		m.ScrollOffset = 0
		return m, siCmd
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
					cmd = m.triggerDebouncedSave() // 使用防抖保存
				}
				m.InputMode = false
				m.StatusMsg = "Comment saved!"
				return m, cmd

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
				// 退出前强制立即保存一次，防止防抖还没触发就退出了
				m.saveStateImmediate()
				// 直接 Quit，屏幕恢复由 WithAltScreen 接管
				return m, tea.Quit

			// 向上移动光标
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
					m.StatusMsg = "" // 移动光标时清除提示消息
					if m.Cursor < m.ScrollOffset {
						m.ScrollOffset = m.Cursor
					}
				}

			// 向下移动光标
			case "down", "j":
				// 限制光标不能超过文件树的总行数
				// 我们需要计算一下当前可见的总节点数
				totalNodes := m.countVisibleNodes(m.RootNode.Children)
				if m.Cursor < totalNodes-1 {
					m.Cursor++
					m.StatusMsg = "" // 移动光标时清除提示消息
					vpHeight := m.viewportHeight()
					if m.Cursor >= m.ScrollOffset+vpHeight {
						m.ScrollOffset = m.Cursor - vpHeight + 1
					}
				}

			// 空格键折叠/展开
			case " ":
				idx := 0
				// 传入 idx 指针，在递归中寻找当前光标对应的节点
				// 如果发生状态改变，触发保存
				if m.toggleNode(m.RootNode.Children, &idx) {
					cmd = m.triggerDebouncedSave() // 使用防抖
				}

			// 回车键隐藏/显示
			case "enter":
				idx := 0
				// 如果发生状态改变，触发保存
				if m.toggleHidden(m.RootNode.Children, &idx) {
					cmd = m.triggerDebouncedSave() // 使用防抖
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

			// 's' 键保存为文本文件
			case "s":
				output := m.generateTreeOutput()
				filename := "gentr_output.txt"
				err := os.WriteFile(filename, []byte(output), 0644)
				if err != nil {
					m.StatusMsg = "Error saving file: " + err.Error()
				} else {
					m.StatusMsg = fmt.Sprintf("Saved to %s", filename)
				}
				return m, tea.Tick(time.Millisecond, func(t time.Time) tea.Msg { return nil })

			// 'p' 键保存两套主题的 SVG 图片
			case "p":
				err1 := m.saveThemeSVG(DarkTheme, "gentr_dark.svg")
				err2 := m.saveThemeSVG(LightTheme, "gentr_light.svg")

				if err1 != nil || err2 != nil {
					m.StatusMsg = "Error saving SVG!"
				} else {
					m.StatusMsg = "Saved gentr_dark.svg & gentr_light.svg"
				}
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

			// 按 '/' 进入搜索模式
			case "/":
				m.SearchMode = true
				m.SearchInput.Focus()
				return m, textinput.Blink

			// 在导航模式按 Esc 清空搜索结果，也退出 Git 模式
			case "esc":
				if m.SearchInput.Value() != "" {
					m.SearchInput.SetValue("")
					m.Cursor = 0 // 重置光标
					m.ScrollOffset = 0
				}
				// 退出 Git 模式
				if m.GitMode {
					m.GitMode = false
					m.Cursor = 0
					m.ScrollOffset = 0
					m.StatusMsg = "Git Filter: OFF"
				}

			// 按 'g' 切换 Git 模式
			case "g":
				m.GitMode = !m.GitMode
				m.Cursor = 0 // 列表变了，重置光标
				m.ScrollOffset = 0
				if m.GitMode {
					m.StatusMsg = "Git Filter: ON (Showing changed files)"
				} else {
					m.StatusMsg = "Git Filter: OFF"
				}
			}
		}
	}
	// 返回 cmd，因为防抖逻辑可能会产生 cmd
	return m, cmd
}

// 触发防抖保存：更新 Tag，并返回一个延时指令
func (m *MainModel) triggerDebouncedSave() tea.Cmd {
	m.SaveTag++ // 版本号 +1
	currentTag := m.SaveTag

	// 返回一个延时 600ms 的指令
	// 如果 600ms 内用户又操作了，m.SaveTag 会继续增加
	// 等这个 Tick 回来时，它的 currentTag 就不等于 m.SaveTag 了，就不会执行保存
	return tea.Tick(600*time.Millisecond, func(t time.Time) tea.Msg {
		return SaveMsg{Tag: currentTag}
	})
}

// saveStateImmediate 立即保存当前状态到 .gentr.json (原 saveState)
func (m MainModel) saveStateImmediate() {
	cwd, _ := os.Getwd()
	_ = core.SaveConfig(cwd, m.RootNode)
}

// shouldShow 判断节点是否应该在当前过滤器(Search && Git)下显示
func (m MainModel) shouldShow(node *model.Node) bool {
	// 1. 搜索词检查
	searchTerm := m.SearchInput.Value()
	matchesSearch := true
	if searchTerm != "" {
		matchesSearch = m.doesNodeMatch(node, strings.ToLower(searchTerm))
	}

	// 2. Git 状态检查
	matchesGit := true
	if m.GitMode {
		matchesGit = m.doesNodeMatchGit(node)
	}

	// 必须同时满足（交集）
	return matchesSearch && matchesGit
}

// doesNodeMatch 递归搜索匹配检查
func (m MainModel) doesNodeMatch(node *model.Node, term string) bool {
	// 1. 检查自己
	if strings.Contains(strings.ToLower(node.Name), term) {
		return true
	}
	// 2. 检查子节点
	// 判断“显不显示”只要有一个子节点匹配就行
	if node.IsDir {
		for _, child := range node.Children {
			if m.doesNodeMatch(child, term) {
				return true
			}
		}
	}
	return false
}

// doesNodeMatchGit 递归 Git 匹配检查
func (m MainModel) doesNodeMatchGit(node *model.Node) bool {
	// 如果自己有状态 (M or A)，匹配
	if node.GitStatus != "" {
		return true
	}
	// 如果是文件夹，检查子节点有没有变的
	if node.IsDir {
		for _, child := range node.Children {
			if m.doesNodeMatchGit(child) {
				return true
			}
		}
	}
	return false
}

// View 渲染终端上的界面
func (m MainModel) View() string {
	if m.Quitting {
		return "" // 退出时返回空字符串，避免屏幕闪烁
	}

	topContent := ""

	// 1. 更新提醒 (最高优先级)
	if m.UpdateAvailable {
		instruction := ""
		if runtime.GOOS == "windows" {
			instruction = "Download at github.com/DoraleCitrus/gentr/releases"
		} else {
			instruction = "Run: curl -sfL .../install.sh | sh"
		}

		msg := fmt.Sprintf("Update Available: %s -> %s | %s", m.CurrentVersion, m.LatestVersion, instruction)

		// 确保横幅占满宽度
		topContent += updateBannerStyle.Width(m.Width).Render(msg) + "\n"
	}

	// 2. 警告条逻辑
	if m.LimitWarning {
		msg := "[!] Safety Limit Reached: Only showing first 5000 files / 10 levels deep."
		topContent += warningStyle.Width(m.Width).Render(msg) + "\n"
	}

	// 3. 标题
	header := fmt.Sprintf("Project: %s\n", m.RootNode.Name)
	topContent += header

	// 递归渲染文件树
	index := 0
	// forceHidden 参数，初始为 false
	fullTreeView := m.renderChildren(m.RootNode.Children, "", &index, false)

	// 处理滚动逻辑
	treeLines := strings.Split(fullTreeView, "\n")
	// 去除最后可能产生的空行
	if len(treeLines) > 0 && treeLines[len(treeLines)-1] == "" {
		treeLines = treeLines[:len(treeLines)-1]
	}

	vpHeight := m.viewportHeight()
	start := m.ScrollOffset
	end := start + vpHeight

	// 边界检查
	if start < 0 {
		start = 0
	}
	if start > len(treeLines) {
		start = len(treeLines)
	}
	if end > len(treeLines) {
		end = len(treeLines)
	}

	visibleLines := treeLines[start:end]
	treeView := strings.Join(visibleLines, "\n")

	// 底部区域逻辑：根据模式切换显示内容
	bottomBar := ""

	if m.InputMode {
		// 1. 如果在输入模式，显示输入框
		bottomBar = fmt.Sprintf("\nAdding comment for selected file:\n%s\n(Enter to save, Esc to cancel)", m.TextInput.View())
	} else if m.SearchMode {
		// 2. 如果在搜索模式，显示搜索框
		bottomBar = fmt.Sprintf("\n%s\n(Enter to view, Esc to cancel)", m.SearchInput.View())
	} else {
		// 3. 如果在导航模式，显示状态栏 + 帮助
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

		// 根据是否是 Git 模式切换状态栏颜色
		currentStatusBarStyle := statusBarStyle
		if m.GitMode {
			currentStatusBarStyle = gitStatusBarStyle
		}
		statusBar := currentStatusBarStyle.Width(m.Width).Render(statusText)

		// 提示文案
		filterHint := ""
		if m.SearchInput.Value() != "" {
			filterHint += " [Esc] Clear Search"
		}
		// Git 模式提示
		if m.GitMode {
			filterHint += " [g] All Files"
		} else {
			filterHint += " [g] Git Changes"
		}

		// 帮助文案
		help := fmt.Sprintf("\n[Spc] Toggle  [Ent] Hide/Show  [i] Comment  [/] Search  %s\n[c] Copy  [s] Save Txt  [p] Save SVG  [q] Quit", filterHint)
		bottomBar = statusBar + help
	}

	result := topContent + treeView + "\n" + bottomBar

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

	// 预先过滤出需要显示的子节点
	var visibleChildren []*model.Node
	for _, child := range children {
		if m.shouldShow(child) {
			visibleChildren = append(visibleChildren, child)
		}
	}

	for i, child := range visibleChildren {
		isLast := i == len(visibleChildren)-1

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

		// Git 颜色逻辑
		// 优先级：光标 > 搜索匹配 > Git状态 > 普通
		// 如果不在光标上，且没有被隐藏
		if *index != m.Cursor && !isNodeHidden {
			if child.GitStatus == "M" {
				style = gitModifiedStyle
			} else if child.GitStatus == "A" {
				style = gitAddedStyle
			}

			// 搜索高亮逻辑 (保留)
			term := m.SearchInput.Value()
			if term != "" && strings.Contains(strings.ToLower(child.Name), strings.ToLower(term)) {
				style = searchMatchStyle
			}
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

		// 构造 Git 标记
		gitMark := ""
		if child.GitStatus == "M" {
			gitMark = " [M]"
		} else if child.GitStatus == "A" {
			gitMark = " [+]"
		}

		// 处理注释的显示逻辑
		annotationStr := ""
		if child.Annotation != "" {
			annotationStr = fmt.Sprintf("  # %s", child.Annotation)
		}

		// 拼接顺序：文件名 + Git标记 + 注释
		totalContent := displayName + gitMark + annotationStr

		// 增加对极小宽度的判断，防止 availableWidth < 0 导致 crash
		if availableWidth <= 1 {
			displayName = "" // 空间太小，直接不显示
			annotationStr = ""
			gitMark = "" // [新增]
		} else {
			// 计算总内容宽度 (名字 + Git标记 + 注释)
			totalWidth := lipgloss.Width(totalContent)

			// 如果总宽度超过可用空间，需要截断
			if totalWidth > availableWidth {
				// 这里的截断策略：优先保证文件名，然后是 Git 标记，最后是注释
				// 为了简化 MVP，我们直接截断 annotationStr
				// 重新计算除注释外的基础宽度
				baseLen := lipgloss.Width(displayName + gitMark)
				if baseLen >= availableWidth {
					// 空间极其紧张，只显示名字
					annotationStr = ""
					gitMark = ""
					runesName := []rune(displayName)
					if availableWidth-1 > 0 {
						displayName = string(runesName[:availableWidth-1]) + "…"
					}
				} else {
					// 截断注释
					remain := availableWidth - baseLen
					runesAnno := []rune(annotationStr)
					if remain > 1 && remain < len(runesAnno) {
						annotationStr = string(runesAnno[:remain-1]) + "…"
					} else {
						annotationStr = ""
					}
				}
			}
		}

		// 渲染 Git 标记的样式
		gitMarkStyle := normalStyle
		if !isNodeHidden {
			if child.GitStatus == "M" {
				gitMarkStyle = gitModifiedStyle
			} else if child.GitStatus == "A" {
				gitMarkStyle = gitAddedStyle
			}
		} else {
			gitMarkStyle = hiddenStyle
		}

		// 拼接字符串：光标指示器 + 缩进 + 连接符 + 文件名 + [Git标记] + [注释]
		line := fmt.Sprintf("%s%s%s%s%s%s%s",
			cursorIndicator,
			dimmedStyle.Render(prefix),
			dimmedStyle.Render(connector),
			icon,
			style.Render(displayName),
			gitMarkStyle.Render(gitMark), // 渲染 Git 标记
			annotationStyle.Render(annotationStr),
		)
		sb.WriteString(line + "\n")

		// 处理完一行, 递增 index
		*index++

		// 如果是文件夹且有子节点，递归渲染其子节点
		// 强制展开逻辑：搜索 或 Git模式下都强制展开
		shouldExpand := !child.Collapsed
		if m.SearchInput.Value() != "" || m.GitMode {
			shouldExpand = true // 强制展开
		}

		if child.IsDir && len(child.Children) > 0 && shouldExpand {
			// 计算新的前缀
			// 如果当前节点是最后一个，那子节点的缩进就是空格 "    "
			// 如果当前节点不是最后一个，那子节点的缩进还需要竖线 "│   " 来连接下面的兄弟节点
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
		// 过滤
		if !m.shouldShow(child) {
			continue
		}

		count++

		// 强制展开逻辑
		shouldExpand := !child.Collapsed
		if m.SearchInput.Value() != "" || m.GitMode {
			shouldExpand = true
		}

		// 只有没折叠的目录 (或搜索/Git时)，才把它的子节点算进总数里
		if child.IsDir && shouldExpand {
			count += m.countVisibleNodes(child.Children)
		}
	}
	return count
}

// toggleNode 找到光标位置的节点并切换状态
func (m MainModel) toggleNode(children []*model.Node, index *int) bool {
	for _, child := range children {
		// 过滤
		if !m.shouldShow(child) {
			continue
		}

		if *index == m.Cursor {
			if child.IsDir {
				child.Collapsed = !child.Collapsed
			}
			return true
		}
		*index++

		// 强制展开逻辑
		shouldExpand := !child.Collapsed
		if m.SearchInput.Value() != "" || m.GitMode {
			shouldExpand = true
		}

		if child.IsDir && shouldExpand {
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
		// 过滤
		if !m.shouldShow(child) {
			continue
		}

		if *index == m.Cursor {
			child.Hidden = !child.Hidden
			return true
		}
		*index++

		// 强制展开逻辑
		shouldExpand := !child.Collapsed
		if m.SearchInput.Value() != "" || m.GitMode {
			shouldExpand = true
		}

		if child.IsDir && shouldExpand {
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
		// 过滤
		if !m.shouldShow(child) {
			continue
		}

		if *index == m.Cursor {
			return child
		}
		*index++

		// 强制展开逻辑
		shouldExpand := !child.Collapsed
		if m.SearchInput.Value() != "" || m.GitMode {
			shouldExpand = true
		}

		if child.IsDir && shouldExpand {
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
	// 增加 shouldShow 过滤，确保导出的内容和看到的搜索结果一致
	var visibleChildren []*model.Node
	for _, child := range children {
		if !child.Hidden && m.shouldShow(child) {
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

		// 根据 GitMode 决定是否追加 Git 标记
		gitSuffix := ""
		if m.GitMode {
			if child.GitStatus == "M" {
				gitSuffix = " [M]"
			} else if child.GitStatus == "A" {
				gitSuffix = " [+]"
			}
		}

		// 输出行：前缀 + 连接线 + 文件名 + [Git标记] + [注释]
		line := fmt.Sprintf("%s%s%s%s", prefix, connector, child.Name, gitSuffix)

		if child.Annotation != "" {
			// 导出时的注释格式，用空格对齐
			line += fmt.Sprintf("  # %s", child.Annotation)
		}
		sb.WriteString(line + "\n")

		// 递归处理子文件夹
		// 强制展开逻辑
		shouldExpand := !child.Collapsed
		if m.SearchInput.Value() != "" || m.GitMode {
			shouldExpand = true
		}

		if child.IsDir && len(child.Children) > 0 && shouldExpand {
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			sb.WriteString(m.generateChildrenText(child.Children, childPrefix))
		}
	}
	return sb.String()
}

// 核心重构：SVG 渲染引擎

// saveThemeSVG 负责生成带主题的 SVG
func (m MainModel) saveThemeSVG(theme Theme, filename string) error {
	lineHeight := 24 // 增加行高，更宽松

	// 1. 生成内容 (XML 格式)
	content, lineCount, maxCharWidth := m.generateSVGContent(m.RootNode, theme)

	// 2. 计算画布
	width := (maxCharWidth * 10) + 60 // 10px per char + padding
	if width < 600 {
		width = 600
	} // 最小宽度
	height := (lineCount * lineHeight) + 60

	var sb strings.Builder
	// Header
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, width, height))
	// Background
	sb.WriteString(fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s" />`, theme.BgColor))

	// 字体栈：显式声明中文字体
	fontFamily := "'Consolas', 'Monaco', 'Microsoft YaHei', 'PingFang SC', 'WenQuanYi Micro Hei', monospace"

	// Style
	sb.WriteString(fmt.Sprintf(`<style>
		text { 
			font-family: %s; 
			font-size: 14px; 
			white-space: pre; 
		}
		.tree { fill: %s; }
		.text { fill: %s; }
		.folder { fill: %s; font-weight: bold; }
		.comment { fill: %s; font-style: italic; }
		.git-mod { fill: %s; font-weight: bold; }
		.git-add { fill: %s; font-weight: bold; }
	</style>`,
		fontFamily,
		theme.TreeColor, theme.TextColor, theme.FolderColor, theme.CommentColor, theme.GitModColor, theme.GitAddColor))

	// Padding Container (Translate)
	sb.WriteString(`<g transform="translate(30, 40)">`) // 左上角留白
	sb.WriteString(content)
	sb.WriteString(`</g></svg>`)

	return os.WriteFile(filename, []byte(sb.String()), 0644)
}

// generateSVGContent 递归生成 SVG 内容
func (m MainModel) generateSVGContent(root *model.Node, theme Theme) (string, int, int) {
	var sb strings.Builder
	lineIndex := 0
	maxLen := 0

	// 根节点，先递增行号
	lineIndex++
	sb.WriteString(fmt.Sprintf(`<text x="0" y="%d" class="folder">%s</text>`, lineIndex*24-24, escapeXML(root.Name)))
	if len(root.Name) > maxLen {
		maxLen = len(root.Name)
	}

	// 递归
	childContent, _, childMax := m.writeSVGRecursive(root.Children, "", theme, &lineIndex)
	sb.WriteString(childContent)

	if childMax > maxLen {
		maxLen = childMax
	}
	return sb.String(), lineIndex, maxLen
}

// writeSVGRecursive 递归生成 XML 标签
func (m MainModel) writeSVGRecursive(children []*model.Node, prefix string, theme Theme, lineIndex *int) (string, int, int) {
	var sb strings.Builder
	maxLen := 0
	lineHeight := 24

	var visibleChildren []*model.Node
	for _, child := range children {
		if !child.Hidden && m.shouldShow(child) {
			visibleChildren = append(visibleChildren, child)
		}
	}

	for i, child := range visibleChildren {
		*lineIndex = *lineIndex + 1 // 先递增行号
		currentY := (*lineIndex - 1) * lineHeight

		isLast := i == len(visibleChildren)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// 开始一行
		sb.WriteString(fmt.Sprintf(`<text x="0" y="%d">`, currentY))

		// 1. 树形线条
		sb.WriteString(fmt.Sprintf(`<tspan class="tree">%s%s</tspan>`, escapeXML(prefix), connector))

		// 2. 图标 (可选，为了美观这里统一用箭头或留空)
		icon := ""
		if child.IsDir {
			icon = "▼ "
		}
		if icon != "" {
			sb.WriteString(fmt.Sprintf(`<tspan class="tree">%s</tspan>`, icon))
		}

		// 3. 文件名 (根据 Git 状态变色)
		nameClass := "text"
		if child.IsDir {
			nameClass = "folder"
		}
		if m.GitMode {
			if child.GitStatus == "M" {
				nameClass = "git-mod"
			}
			if child.GitStatus == "A" {
				nameClass = "git-add"
			}
		}
		sb.WriteString(fmt.Sprintf(`<tspan class="%s">%s</tspan>`, nameClass, escapeXML(child.Name)))

		// 4. Git 标记 (仅在 Git 模式下)
		if m.GitMode {
			mark := ""
			markClass := ""
			if child.GitStatus == "M" {
				mark = " [M]"
				markClass = "git-mod"
			} else if child.GitStatus == "A" {
				mark = " [+]"
				markClass = "git-add"
			}
			if mark != "" {
				sb.WriteString(fmt.Sprintf(`<tspan class="%s">%s</tspan>`, markClass, mark))
			}
		}

		// 5. 注释
		if child.Annotation != "" {
			sb.WriteString(fmt.Sprintf(`<tspan class="comment">  # %s</tspan>`, escapeXML(child.Annotation)))
		}

		sb.WriteString(`</text>`)

		// 计算粗略宽度
		rowLen := len(prefix) + 4 + len(child.Name) + len(child.Annotation) + 5
		if rowLen > maxLen {
			maxLen = rowLen
		}

		// 递归子节点
		shouldExpand := !child.Collapsed
		if m.SearchInput.Value() != "" || m.GitMode {
			shouldExpand = true
		}

		if child.IsDir && len(child.Children) > 0 && shouldExpand {
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			cContent, _, cMax := m.writeSVGRecursive(child.Children, childPrefix, theme, lineIndex)
			sb.WriteString(cContent)
			if cMax > maxLen {
				maxLen = cMax
			}
		}
	}
	return sb.String(), 0, maxLen
}

// escapeXML 转义 XML 特殊字符
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
