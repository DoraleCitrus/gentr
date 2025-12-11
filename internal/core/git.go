package core

import (
	"bytes"
	"os/exec"
	"path/filepath"
)

// LoadGitStatus 返回一个 map，key 是文件的相对路径，value 是状态代码 ("M", "A", etc.)
func LoadGitStatus(rootPath string) map[string]string {
	statusMap := make(map[string]string)

	// 尝试使用系统安装的 git 命令，因为它能更好地处理配置（如 core.filemode, core.autocrlf）
	// 使用 -z 选项以 NUL 字符分隔输出，避免文件名包含特殊字符的问题
	cmd := exec.Command("git", "status", "--porcelain", "-z")
	cmd.Dir = rootPath
	output, err := cmd.Output()
	if err != nil {
		// 如果执行失败（例如未安装 git 或不是 git 仓库），返回空 map
		return statusMap
	}

	// 解析输出
	// 格式: XY PATH\0 [ORIG_PATH\0]
	entries := bytes.Split(output, []byte{0})
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < 4 { // 至少包含状态(2) + 空格(1) + 路径(1+)
			continue
		}

		// 状态码
		x := entry[0]
		y := entry[1]
		// 路径从索引 3 开始 (跳过 "XY ")
		path := string(entry[3:])

		// 处理重命名 (R)：如果是重命名，下一个 entry 是原始路径，需要跳过
		if x == 'R' {
			i++
		}

		var s string
		// 映射状态
		// ?? -> Untracked -> A
		// A. -> Added -> A
		// .A -> Added -> A
		// M. -> Modified -> M
		// .M -> Modified -> M

		if x == '?' && y == '?' {
			s = "A"
		} else if x == 'A' || y == 'A' {
			s = "A"
		} else if x == 'M' || y == 'M' {
			s = "M"
		}

		if s != "" {
			cleanPath := filepath.ToSlash(path)
			statusMap[cleanPath] = s
		}
	}

	return statusMap
}
