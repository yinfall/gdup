<p align="right">
  <a href="README.md">English</a> | <strong>简体中文</strong>
</p>

# GDUP - Godot Version Manager

GDUP 是一个轻量级、跨平台的 Godot 引擎版本管理器，采用 Go 语言原生开发。它旨在帮助开发者在单一物理机上高效地安装、管理并无缝切换多个 Godot 引擎版本，以满足多项目并行开发的需求。

## ✨ 核心特性

- **项目级版本隔离**：基于目录结构的配置识别能力。通过在项目根目录创建 `.gduprc` 文件，系统可自动为不同项目匹配并启动对应的引擎版本。
- **透明代理架构 (Shim)**：提供可选的全局命令劫持功能。开启后，开发者可在终端直接执行标准 `godot` 命令，GDUP 会以零延迟将其安全代理至目标版本引擎，完全不改变既有开发习惯。
- **零外部依赖**：通过静态编译分发单文件二进制程序，无需预装任何运行时环境（如 Python、Node.js 等）。
- **统一存储与空间复用**：所有下载的引擎版本均集中维护在 `~/.gdup/versions` 目录下。支持通过 `GDUP_CACHE_PATH` 环境变量自定义存储路径，避免磁盘空间浪费。
- **原生跨平台支持**：提供 Windows、macOS (Intel/Apple Silicon) 及 Linux 的原生二进制支持。

---

## 🚀 安装指南

系统将自动下载适用您环境的最新二进制文件，并统一安装至用户目录的 `~/.gdup/bin` 中。

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.ps1 | iex
```

### macOS / Linux (Terminal)
```bash
curl -fsSL https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.sh | bash
```

> **注意**：安装完成后，请务必将 `~/.gdup/bin` （Windows 为 `C:\Users\<您的用户名>\.gdup\bin`）添加至操作系统的全局环境变量 `PATH` 中。

---

## 🛠️ 基本使用

通过 `gdup` 主命令，您可以进行引擎版本的完整生命周期管理：

```bash
gdup releases         # 检索远程可用的 Godot 官方发布版本
gdup install 4.3      # 下载并安装指定的版本
gdup use 4.3          # 在当前工作目录配置使用特定版本 (生成 .gduprc)
gdup list             # 列出本地已安装的所有版本
gdup godot --editor   # 通过 GDUP 显式启动匹配当前项目的 Godot 引擎
```

### 代理环境配置 (Shim)

若希望直接使用原生的 `godot` 命令（而非 `gdup godot`）进行调用，可启用透明代理模式：

```bash
# 安装并启用透明代理
gdup shim install
```
启用后，您可以在任意已配置 `.gduprc` 的项目目录中直接执行 `godot`，系统将在底层自动完成版本路由。

```bash
# 移除透明代理，恢复系统默认环境
gdup shim remove
```

---

## 📚 开发者文档

关于 GDUP 底层的目录嗅探机制、单体应用 (Fat Binary) 架构设计、环境变量配置说明，以及如何参与贡献，请查阅 [开发者与架构文档](docs/development.md)。

## 📄 许可证
本项目采用 [MIT License](LICENSE) 授权协议。
