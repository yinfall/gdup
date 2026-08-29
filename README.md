# GDUP - Godot Version Manager

<p align="center">
  一款极速、纯粹、跨平台的 Godot 引擎版本管理器，采用 Go 语言原生编写。
</p>

## 🌟 为什么选择 GDUP？

- **极速纯净**：单文件免安装，无任何外部依赖，丝滑切换不同项目的 Godot 引擎版本。
- **无痛集成**：原生支持智能代理垫片 (Shim)，终端里盲敲 \godot\ 即可自动适配当前项目版本。
- **完美跨平台**：全面支持 Windows, macOS, 以及 Linux。

---

## 🚀 一键安装

GDUP 会自动安装到你用户目录下的 \~/.gdup/bin\ 中。

### Windows (PowerShell)
\\\powershell
irm https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.ps1 | iex
\\\

### macOS / Linux (终端)
\\\ash
curl -fsSL https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.sh | bash
\\\

⚠️ **重要提示**：安装成功后，请根据终端提示将 \~/.gdup/bin\ （Windows 为 \C:\Users\你的用户名\.gdup\bin\）路径添加到你的系统环境变量 **PATH** 中！

---

## 🛠️ 快速上手

\\\ash
gdup releases         # 看看线上有哪些版本可以下
gdup install 4.3      # 下载并安装 4.3 版本
gdup use 4.3          # 将当前文件夹的项目绑定到 4.3 版本
gdup list             # 查看本地装了哪些版本
gdup godot --editor   # 像平时一样启动 Godot 引擎
\\\

### 🧙‍♂️ 开启“魔法垫片” (推荐)
想要连 \gdup\ 这个前缀都省掉，直接像往常一样使用原生的 \godot\ 命令？
只需执行：
\\\ash
gdup shim install
\\\
现在，你在任何项目里直接敲 \godot\，系统就能智能嗅探并启动该项目对应的正确版本了！

---

## 📚 开发者指南与底层架构

想了解 GDUP 是如何实现 0 延迟启动的？如何修改源码或自定义缓存目录配置？
请参阅我们的 [开发者开发与架构指南](docs/development.md)。

## 📄 开源许可
MIT License
