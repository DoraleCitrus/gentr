package model

//Node 是一个节点，代表目录树中的一个文件或文件夹
type Node struct {
	Name     string  //文件名，e.g. "main.go"
	Path     string  //文件完整路径，e.g. "/usr/local/project/main.go"
	IsDir    bool    //是否为文件夹
	Children []*Node //子节点列表，仅当 IsDir 为 true 时有效

	// 以下字段用于UI交互
	Collapsed  bool   //是否折叠
	Hidden     bool   //是否隐藏(用户手动排除)
	Annotation string //用户注释

	// Git 状态
	// "" = 无变化, "M" = 修改, "A" = 新增/未追踪, "?" = 未知
	GitStatus string
}
