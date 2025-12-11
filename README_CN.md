<div align="center">
  <h1>🌳 Gentr</h1>
  <p>
    <strong>智能、交互式的项目结构树生成器</strong>
  </p>
  <p>
    优雅地浏览、过滤、注释并导出你的项目结构。提供更加可读的git status替代。
  </p>

  <p>
    <a href="README.md"><strong>English Docs</strong></a> |
    <a href="#-安装">安装</a> |
    <a href="#-使用指南">使用指南</a> |
    <a href="#-功能特性">功能特性</a>
  </p>

<a href="https://github.com/DoraleCitrus/gentr/actions"><img src="https://github.com/DoraleCitrus/gentr/actions/workflows/release.yml/badge.svg" alt="Build Status"></a>
<a href="https://goreportcard.com/report/github.com/DoraleCitrus/gentr"><img src="https://goreportcard.com/badge/github.com/DoraleCitrus/gentr" alt="Go Report Card"></a>
<a href="https://github.com/DoraleCitrus/gentr/blob/main/LICENSE"><img src="https://img.shields.io/github/license/DoraleCitrus/gentr" alt="License"></a>
<a href="https://github.com/DoraleCitrus/gentr/releases"><img src="https://img.shields.io/github/v/release/DoraleCitrus/gentr" alt="Latest Release"></a>

</div>

## 📖 简介

**Gentr** 是经典 `tree` 命令的现代化升级版。它是一个基于终端（TUI）的交互式工具，专为编写技术文档、代码审查和项目交接而设计。

使用 Gentr，你可以告别手动调整文本格式的目录树痛苦。你可以交互式地隐藏无关文件、高亮 Git 变更、添加代码注释，并一键导出为 Markdown 或精美的 SVG 图片。

厌倦了 git status 的糟糕可读性？这里也有更加优雅的方式来预览 git 变更。

## ✨ 功能特性

- **🔍 模糊搜索：** 按下 <kbd>/</kbd> 键，瞬间过滤深层文件结构。
- **🐙 Git 集成：** 可视化 `[+]` 新增和 `[M]` 修改的文件。按 <kbd>g</kbd> 键仅显示发生变更的文件树。
- **📝 代码注释：** 按 <kbd>i</kbd> 键为文件添加注释（例如：`# 程序入口`）。注释会自动保存。
- **🖼️ 强大的导出：**
  - 复制 Markdown 到剪贴板 (<kbd>c</kbd>)。
  - 导出 **深色/浅色主题 SVG 图片** (<kbd>p</kbd>)。
  - 保存为 txt 文本文件 (<kbd>s</kbd>)。
- **🛡️ 智能安全：** 默认遵循 `.gitignore` 规则。内置防崩溃保护机制，默认设置深度为 10，节点数为 5000 的上限。可用命令行参数强制无视。

## 🚀 安装

### 一键安装 (Linux & macOS)

```bash
curl -sfL https://raw.githubusercontent.com/DoraleCitrus/gentr/main/install.sh | sh
```

#### **macOS 用户注意**: 如果遇到“无法验证开发者”的安全拦截，请运行以下命令来信任该程序。

```bash
xattr -d com.apple.quarantine $(which gentr)
```

### Windows (手动安装)

前往 [Releases 页面](https://github.com/DoraleCitrus/gentr/releases) 下载最新的 gentr_x.x.x_windows_amd64.zip。

将压缩包解压到一个你不会删除的目录（例如 C:\gentr\）。

配置环境变量 (PATH)：

按 <kbd>Win</kbd> 键，搜索 "环境变量" 并打开。

点击 "环境变量(N)..." 按钮。

在 "用户变量" 列表中找到 Path，选中并点击 "编辑"。

点击 "新建"，将刚才的文件夹路径（如 C:\gentr\）粘贴进去。

一路点击 "确定" 保存。

重启你的终端 (PowerShell / CMD / VS Code) 即可生效。

### 手动下载

请前往 [Releases 页面](https://github.com/DoraleCitrus/gentr/releases) 下载适用于 Windows、macOS 和 Linux 的预编译二进制文件。

### Go 安装 (开发者)

如果你已安装 Go 环境：

```bash
go install github.com/DoraleCitrus/gentr@latest
```

## ⌨️ 使用指南

在项目根目录下运行：

```bash
gentr
# 或者指定路径
gentr -p ./src
```

### 快捷键列表

| 按键                                                  | 功能                        |
| :---------------------------------------------------- | :-------------------------- |
| <kbd>↑</kbd> <kbd>↓</kbd> / <kbd>k</kbd> <kbd>j</kbd> | 移动光标                    |
| <kbd>Space</kbd>                                      | 折叠 / 展开文件夹           |
| <kbd>Enter</kbd>                                      | 隐藏 / 显示 文件 (变灰)     |
| <kbd>/</kbd>                                          | 模糊搜索 (Esc 清除)         |
| <kbd>g</kbd>                                          | 切换 Git 变更过滤器         |
| <kbd>i</kbd>                                          | 添加 / 编辑 注释            |
| <kbd>c</kbd>                                          | 复制 结果到剪贴板           |
| <kbd>p</kbd>                                          | 导出 SVG 图片 (深色 & 浅色) |
| <kbd>s</kbd>                                          | 保存为 .txt 文件            |
| <kbd>q</kbd>                                          | 退出                        |

### 命令行参数

```bash
-p, --path <dir>   指定目标目录 (默认: 当前目录)
-f, --force        强制模式: 无视 .gitignore 和文件数量限制 (危险!)
-v, --version      显示版本信息
-h, --help         显示帮助信息
```

## 💾 持久化配置存储

Gentr 会在你的项目根目录下自动生成一个 `.gentr.json` 文件。它用于存储：

- 被隐藏的文件列表
- 文件夹的折叠状态
- 你编写的自定义注释

**提示：** 将 `.gentr.json` 提交到 Git 仓库，即可与团队成员共享这份文档结构！

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 开源协议

本项目基于 MIT 协议开源 - 详见 [LICENSE](https://github.com/DoraleCitrus/gentr/blob/main/LICENSE) 文件。
