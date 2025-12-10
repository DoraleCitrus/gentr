package core

import (
	"os"
	"path/filepath"

	"github.com/DoraleCitrus/gentr/internal/model"
	ignore "github.com/sabhiram/go-gitignore"
)

// Walk 负责从根目录开始构建树
func Walk(rootPath string) (*model.Node, error) {
	// 寻找并加载根目录底下的.gitignore文件
	ignoreObj, _ := ignore.CompileIgnoreFile(filepath.Join(rootPath, ".gitignore"))
	// 开始递归扫描目录
	return scanDir(rootPath, ignoreObj)
}

// scanDir 递归扫描目录，构建节点树
func scanDir(path string, ignoreObj *ignore.GitIgnore) (*model.Node, error) {
	// 获取文件或文件夹的基础信息
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

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
		// 构建子文件的完整路径
		fullPath := filepath.Join(path, entry.Name())

		// 检查是否被.gitignore忽略
		if ignoreObj != nil && ignoreObj.MatchesPath(entry.Name()) { // ignoreObj 可能为空（如果用户没写 .gitignore），所以要判空
			continue
		}

		// 默认忽略.git
		if entry.Name() == ".git" {
			continue
		}

		// 递归调用
		childNode, err := scanDir(fullPath, ignoreObj)
		if err != nil {
			continue // 遇到错误选择跳过而不是崩溃
		}
		node.Children = append(node.Children, childNode) //把找到的子节点挂载到当前节点下
	}
	return node, nil
}
