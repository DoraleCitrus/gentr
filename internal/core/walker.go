package core

import (
	"os"
	"path/filepath"

	"github.com/DoraleCitrus/gentr/internal/model"
	ignore "github.com/sabhiram/go-gitignore"
)

// 安全限制常量
const (
	MaxFiles = 5000 // 最大文件节点数
	MaxDepth = 10   // 最大递归深度
)

// 计数器
type counter struct {
	count        int
	limitReached bool
}

// Walk 负责从根目录开始构建树
func Walk(rootPath string) (*model.Node, bool, error) {
	// 加载 .gitignore
	ignoreObj, _ := ignore.CompileIgnoreFile(filepath.Join(rootPath, ".gitignore"))

	// 初始化计数器
	c := &counter{count: 0}

	// 开始扫描
	root, err := scanDir(rootPath, ignoreObj, 0, c)

	// 返回结果，同时返回是否触发了限制
	return root, c.limitReached, err
}

// scanDir 递归扫描目录，构建节点树
func scanDir(path string, ignoreObj *ignore.GitIgnore, depth int, c *counter) (*model.Node, error) {
	// 深度熔断检查
	if depth > MaxDepth {
		return nil, nil // 超过深度，不再深入，直接返回 nil，会被上层过滤掉
	}

	// 数量熔断检查
	if c.count >= MaxFiles {
		c.limitReached = true // 标记触发限制
		return nil, nil
	}

	// 获取文件或文件夹的基础信息
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// 计数
	c.count++

	// 创建当前节点
	node := &model.Node{
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
	}

	// 如果不是文件夹，返回节点
	if !info.IsDir() {
		return node, nil
	}

	// 是文件夹则需要扫描子内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// git 目录硬编码忽略
		if entry.Name() == ".git" {
			continue
		}

		// .gitignore 检查
		if ignoreObj != nil && ignoreObj.MatchesPath(entry.Name()) {
			continue
		}

		// 构建子文件的完整路径
		fullPath := filepath.Join(path, entry.Name())

		// 递归调用 (深度 + 1)
		childNode, err := scanDir(fullPath, ignoreObj, depth+1, c)
		if err != nil {
			continue // 遇到错误选择跳过而不是崩溃
		}

		// childNode 为 nil 时不添加
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}

		// 数量限制优化，提前跳出
		if c.limitReached {
			break
		}
	}
	return node, nil
}
