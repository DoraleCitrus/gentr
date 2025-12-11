package core

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// LoadGitStatus 返回一个 map，key 是文件的相对路径，value 是状态代码 ("M", "A", etc.)
func LoadGitStatus(rootPath string) map[string]string {
	statusMap := make(map[string]string)

	// 打开仓库
	r, err := git.PlainOpen(rootPath)
	if err != nil {
		return statusMap // 不是 git 仓库，直接返回空 map
	}

	// 获取工作区
	w, err := r.Worktree()
	if err != nil {
		return statusMap
	}

	// 获取状态
	status, err := w.Status()
	if err != nil {
		return statusMap
	}

	// 转换 map
	for path, fileStatus := range status {
		// go-git 返回的状态代码是 ASCII 字符
		// Worktree 表示工作区的状态，Staging 表示暂存区的状态

		s := ""
		// 如果是 Untracked (?) 或者 Added (A)
		if fileStatus.Worktree == git.Untracked || fileStatus.Staging == git.Added || fileStatus.Worktree == git.Added {
			s = "A"
		} else if fileStatus.Worktree == git.Modified || fileStatus.Staging == git.Modified {
			s = "M"
		}

		if s != "" {
			// 统一路径分隔符，防止 Windows 下匹配失败
			cleanPath := filepath.ToSlash(path)
			statusMap[cleanPath] = s
		}
	}

	return statusMap
}
