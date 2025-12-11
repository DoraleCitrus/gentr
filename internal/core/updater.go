package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GitHub Release API 结构体
type GitHubRelease struct {
	TagName string `json:"tag_name"` // e.g., "v1.0.1"
	HTMLURL string `json:"html_url"`
}

// CheckForUpdates 检查是否有新版本
// 返回: (是否有更新, 最新版本号, 错误)
func CheckForUpdates(currentVersion string) (bool, string, error) {
	// 设置超时，防止网络卡死
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	url := "https://api.github.com/repos/DoraleCitrus/gentr/releases/latest"
	resp, err := client.Get(url)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", err
	}

	// 简单的版本比较逻辑("vX.Y.Z" 格式)
	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	// 如果字符串不相等，且 API 返回的不为空，我们认为有更新
	if latest != "" && latest != current {
		return true, release.TagName, nil
	}

	return false, "", nil
}
