# GDUP - Godot Version Manager

GDUP 是一款极速、跨平台的 Godot 引擎版本管理器，采用 Go 语言编写。它可以帮助你在不同的 Godot 版本之间无缝切换，并支持极客体验的“透明垫片 (Shim)” 架构。

## 🌟 核心特性

- **独立 CLI 设计**：原生的 \gdup\ 命令，告别名称冲突，逻辑纯净。
- **神级透明垫片 (Opt-in Shim)**：支持 \gdup shim install\ 一键生成自代理垫片。开启后，在终端盲敲 \godot\ 即可自动智能识别当前项目的版本并极速启动引擎！
- **目录级版本隔离**：支持通过项目根目录下的 \.gduprc\ 配置不同项目。如果没有，它会自动向上回溯父级目录，或最终匹配家目录的全局配置。
- **一键极速安装**：自动连接 GitHub 获取最新版本，支持跨平台原生编译（Windows / macOS / Linux）。
- **完全绿色**：引擎全部集中存放在 \~/.gdup/versions\ 下，不乱写系统注册表，随删随走。

---

## 🚀 安装指南

系统会自动下载最新版本，并将其安装到 \~/.gdup/bin\ 目录下。

### Windows (PowerShell)
`powershell
irm https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.ps1 | iex
`

### macOS / Linux (终端)
`ash
curl -fsSL https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.sh | bash
`

> **注意：** 安装完成后，请记得将 \~/.gdup/bin\ (或 Windows 对应的 \C:\Users\你的用户名\.gdup\bin\) 路径添加到你的系统环境变量 \PATH\ 中！

---

## 🛠️ 基本用法

\\\ash
# 查看所有可下载的 Godot 官方版本
gdup releases

# 下载并安装某个特定版本
gdup install 4.3

# 查看本地已经安装的版本
gdup list

# 在当前目录下配置使用特定版本 (会生成 .gduprc)
gdup use 4.3

# 类似 fvm，直接通过代理运行 Godot 引擎并透传参数
gdup godot --editor
\\\

---

## 🧙‍♂️ 魔法功能：透明垫片 (Shim)

如果你不想每次都敲 \gdup godot\，你可以开启极具魔法感的透明垫片功能！

\\\ash
# 开启透明垫片
gdup shim install
\\\
开启后，GDUP 会自动复制自身并生成一个极速 \godot\ 执行环境。你现在可以直接在终端输入：
\\\ash
godot --editor
\\\
它会在不到 1 毫秒内嗅探当前目录下的 \.gduprc\，并**100% 透明**地把参数转发给真实的引擎。

如果不再需要垫片，随时关闭：
\\\ash
gdup shim remove
\\\

---

## ⚙️ 高级配置 (环境变量)

- \GDUP_CACHE_PATH\: 如果你不想把几十 GB 的 Godot 引擎装在 C 盘或家目录，可以配置此环境变量 (例如 \D:\GodotCache\)，所有的引擎版本会自动下载并存放到 \GDUP_CACHE_PATH/versions\ 中。

---

## 📄 开源许可

本项目采用 [MIT License](LICENSE) 许可开源。
