package main

import (
	"bufio" // 用于读取用户输入
	"flag"  // 用于解析命令行参数
	"fmt"
	"os"
	"path/filepath" // 用于处理路径
	"strings"       // 用于处理字符串

	"github.com/DoraleCitrus/gentr/internal/core"
	"github.com/DoraleCitrus/gentr/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// 定义版本号
const Version = "1.0.2"

func main() {
	// 定义命令行参数 Flags
	var (
		pathFlag    string
		showVersion bool
		forceMode   bool
	)

	// 自定义帮助信息 (-h / --help)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gentr - A smart project tree generator CLI tool.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  gentr [flags] [path]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -p, --path <dir>   Target directory path (default: current directory)\n")
		fmt.Fprintf(os.Stderr, "  -f, --force        Force mode: Ignore .gitignore and file limits (Dangerous!)\n")
		fmt.Fprintf(os.Stderr, "  -v, --version      Show version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help         Show this help message\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  gentr\n")
		fmt.Fprintf(os.Stderr, "  gentr src/\n")
		fmt.Fprintf(os.Stderr, "  gentr -p ../other-project\n")
	}

	// 绑定 Flags
	flag.StringVar(&pathFlag, "p", "", "Target directory path")
	flag.StringVar(&pathFlag, "path", "", "Target directory path")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&forceMode, "f", false, "Force mode")
	flag.BoolVar(&forceMode, "force", false, "Force mode")

	flag.Parse()

	// 处理版本号
	if showVersion {
		fmt.Printf("gentr version %s\n", Version)
		os.Exit(0)
	}

	// 解析目标路径 (优先级: -p > 位置参数 > 当前目录)
	targetPath := "."
	if pathFlag != "" {
		targetPath = pathFlag
	} else if flag.NArg() > 0 {
		targetPath = flag.Arg(0)
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// 验证路径有效性
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		fmt.Printf("[Error] Path '%s' is not a valid directory.\n", absPath)
		os.Exit(1)
	}

	// 处理配置选项与 Force 模式交互
	opts := core.DefaultOptions()

	if forceMode {
		fmt.Println("⚠️  WARNING: Force Mode Enabled")
		fmt.Println("------------------------------------------------")
		fmt.Println("This will ignore .gitignore rules and remove all file limits.")
		fmt.Println("Running this on large directories (like node_modules or system roots) may cause:")
		fmt.Println("  - High memory usage")
		fmt.Println("  - Application freeze")
		fmt.Println("")
		fmt.Print("Are you sure you want to proceed? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "y" && input != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}

		opts = core.ForceOptions()
		fmt.Println("[!]Starting in Force Mode...")
	}

	// 扫描文件
	rootNode, limitReached, err := core.Walk(absPath, opts)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	// 如果有则加载持久化配置
	// 会修改 rootNode 里的 Annotation/Hidden/Collapsed 状态
	// 使用 absPath 作为配置加载路径
	core.LoadConfig(absPath, rootNode)

	// 初始化 UI 模型：传入 Version 以便进行更新检查
	initialModel := ui.InitialModel(rootNode, limitReached, Version)

	// 创建 Bubble Tea 程序并运行
	// 使用 tea.WithAltScreen() 确保程序由框架接管全屏模式
	// 这样退出时框架会自动恢复终端状态，解决无法打字的问题
	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("[ERROR]: %v", err)
		os.Exit(1)
	}
}
