package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/DoraleCitrus/gentr/internal/model"
)

const ConfigFileName = ".gentr.json"

// NodeConfig 定义了每个节点需要持久化的状态
type NodeConfig struct {
	Annotation string `json:"annotation,omitempty"` // omitempty: 如果为空就不存，节省空间
	Collapsed  bool   `json:"collapsed,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
}

// ConfigFile 是最终存入 JSON 的结构
type ConfigFile struct {
	// Key 是文件的相对路径 (例如 "cmd/main.go")
	Nodes map[string]NodeConfig `json:"nodes"`
}

// LoadConfig 读取配置文件并将其应用到现有的树结构上
func LoadConfig(rootPath string, rootNode *model.Node) {
	configPath := filepath.Join(rootPath, ConfigFileName)

	// 1. 读取文件
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return // 文件不存在，直接跳过
	}
	if err != nil {
		return // 读取错误也跳过，不做处理
	}

	// 2. 解析 JSON
	var config ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}

	// 3. 递归遍历树，应用配置
	applyConfig(rootNode, rootPath, config.Nodes)
}

func applyConfig(node *model.Node, rootPath string, configMap map[string]NodeConfig) {
	// 计算当前节点的相对路径
	// node.Path 是绝对路径，我们需要把它变成相对于项目根目录的路径
	relPath, err := filepath.Rel(rootPath, node.Path)
	if err == nil {
		// 为了跨平台兼容，统一把路径分隔符转为 "/"
		relPath = filepath.ToSlash(relPath)

		if conf, ok := configMap[relPath]; ok {
			node.Annotation = conf.Annotation
			node.Collapsed = conf.Collapsed
			node.Hidden = conf.Hidden
		}
	}

	for _, child := range node.Children {
		applyConfig(child, rootPath, configMap)
	}
}

// SaveConfig 收集当前树的状态并写入文件
func SaveConfig(rootPath string, rootNode *model.Node) error {
	configPath := filepath.Join(rootPath, ConfigFileName)

	config := ConfigFile{
		Nodes: make(map[string]NodeConfig),
	}

	// 1. 递归收集状态
	collectConfig(rootNode, rootPath, config.Nodes)

	// 2. 序列化为 JSON (Indent 让文件人类可读)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 3. 写入文件
	return os.WriteFile(configPath, data, 0644)
}

func collectConfig(node *model.Node, rootPath string, configMap map[string]NodeConfig) {
	// 只有当节点有状态改变时才保存（节省空间）
	if node.Annotation != "" || node.Collapsed || node.Hidden {
		relPath, err := filepath.Rel(rootPath, node.Path)
		if err == nil {
			relPath = filepath.ToSlash(relPath)
			configMap[relPath] = NodeConfig{
				Annotation: node.Annotation,
				Collapsed:  node.Collapsed,
				Hidden:     node.Hidden,
			}
		}
	}

	for _, child := range node.Children {
		collectConfig(child, rootPath, configMap)
	}
}
