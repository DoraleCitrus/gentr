<div align="center">
  <h1>🌳 Gentr</h1>
  <p>
    <strong>The intelligent, interactive project tree generator for developers.</strong>
  </p>
  <p>
    Navigate, filter, annotate, and export your project structure with style.And git status in a readable way.
  </p>

  <p>
    <a href="README_CN.md"><strong>中文文档</strong></a> |
    <a href="#-installation">Installation</a> |
    <a href="#-usage">Usage</a> |
    <a href="#-features">Features</a>
  </p>

<a href="https://github.com/DoraleCitrus/gentr/actions"><img src="https://github.com/DoraleCitrus/gentr/actions/workflows/release.yml/badge.svg" alt="Build Status"></a>
<a href="https://goreportcard.com/report/github.com/DoraleCitrus/gentr"><img src="https://goreportcard.com/badge/github.com/DoraleCitrus/gentr" alt="Go Report Card"></a>
<a href="https://github.com/DoraleCitrus/gentr/blob/main/LICENSE"><img src="https://img.shields.io/github/license/DoraleCitrus/gentr" alt="License"></a>
<a href="https://github.com/DoraleCitrus/gentr/releases"><img src="https://img.shields.io/github/v/release/DoraleCitrus/gentr" alt="Latest Release"></a>

</div>

## 📖 Introduction

**Gentr** takes the classic `tree` command and gives it superpowers. It is a TUI (Terminal User Interface) tool designed for documentation, code reviews, and project onboarding.

Stop manually formatting text trees for your PRs. With Gentr, you can interactively hide clutter, highlight git changes, add comments, and export beautiful SVG images.

Tired of the poor readability of git status? We can make it more elegant for you too.

## ✨ Features

- **🔍 Fuzzy Search:** Instantly filter deep file structures by pressing <kbd>/</kbd>.
- **🐙 Git Awareness:** Visualize `[+]` added and `[M]` modified files. Filter to show _only_ changed files with <kbd>g</kbd>.
- **📝 Annotations:** Press <kbd>i</kbd> to add comments to files (e.g., `# Entry Point`). Comments are auto-saved.
- **🖼️ Beautiful Exports:**
  - Copy Markdown to clipboard (<kbd>c</kbd>).
  - Export **Dark/Light Theme SVGs** (<kbd>p</kbd>).
  - Save to text file (<kbd>s</kbd>).
- **🛡️ Smart & Safe:** Respects `.gitignore` by default. Includes safety limits for large directories (configurable).

## 🚀 Installation

### One-line Install (Linux & macOS)

```bash
curl -sfL https://raw.githubusercontent.com/DoraleCitrus/gentr/main/install.sh | sh
```

#### **Note for macOS users**: If you see a security warning, run this command to trust the binary.

```bash
xattr -d com.apple.quarantine $(which gentr)
```

### Windows (Manual Install)

Download the latest gentr_x.x.x_windows_amd64.zip from the [Releases Page](https://github.com/DoraleCitrus/gentr/releases).

Extract the zip file to a permanent location (e.g., C:\gentr\).

Add to PATH (so you can run gentr from any terminal):

Press <kbd>Win</kbd> and search for "env".

Select "Edit the system environment variables".

Click "Environment Variables...".

Under "User variables", find Path and click Edit.

Click New and paste the folder path (e.g., C:\gentr\).

Click OK on all windows.

Restart your terminal (PowerShell / CMD / VS Code) to apply changes.

### Manual Download

Download the pre-compiled binaries for Windows, macOS, and Linux from the [Releases Page](https://github.com/DoraleCitrus/gentr/releases).

### Go Install (Developers)

If you have Go installed:

```bash
go install github.com/DoraleCitrus/gentr@latest
```

## ⌨️ Usage

Run `gentr` in your project root:

```bash
gentr
# Or specify a path
gentr -p ./src
```

### Keybindings

| Key                                                   | Action                           |
| :---------------------------------------------------- | :------------------------------- |
| <kbd>↑</kbd> <kbd>↓</kbd> / <kbd>k</kbd> <kbd>j</kbd> | Move cursor                      |
| <kbd>Space</kbd>                                      | Toggle folder collapse/expand    |
| <kbd>Enter</kbd>                                      | Hide/Show file (Soft delete)     |
| <kbd>/</kbd>                                          | Fuzzy Search (Esc to clear)      |
| <kbd>g</kbd>                                          | Toggle Git Change Filter         |
| <kbd>i</kbd>                                          | Add/Edit Comment                 |
| <kbd>c</kbd>                                          | Copy tree to clipboard           |
| <kbd>p</kbd>                                          | Export SVG images (Dark & Light) |
| <kbd>s</kbd>                                          | Save to .txt file                |
| <kbd>q</kbd>                                          | Quit                             |

### CLI Flags

```bash
-p, --path <dir>   Target directory path (default: current directory)
-f, --force        Force mode: Ignore .gitignore and file limits (Dangerous!)
-v, --version      Show version information
-h, --help         Show help message
```

## 💾 Persistence

Gentr automatically creates a `.gentr.json` file in your project root. This file stores your:

- Hidden files configuration
- Collapsed folder states
- Custom annotations

**Tip:** Commit `.gentr.json` to your repository to share the documentation structure with your team!

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/DoraleCitrus/gentr/blob/main/LICENSE) file for details.
