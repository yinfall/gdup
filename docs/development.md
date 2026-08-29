# GDUP 开发与架构指南

欢迎来到 GDUP 的开发文档！本文档主要面向想要了解底层原理、自行编译或贡献代码的开发者。

## 🏗️ 核心架构设计

GDUP 摒弃了传统的 `.bat` 或 `.sh` 脚本包装方案，采用业界先进的 **“胖二进制 (Fat Binary) / 多面手”** 模式，以实现极致的启动速度和完美的进程信号传递。

### 1. 透明垫片 (Shim) 的自我分身
为了实现最优雅的 `godot` 命令拦截，GDUP 支持 `gdup shim install` 功能。
开启该功能后，`gdup.exe` 会在安装目录 (`~/.gdup/bin`) 下拷贝一份自己，并重命名为 `godot.exe`。

在程序启动时，它会嗅探自身的调用名称 (`os.Args[0]`)：
- **作为 `gdup` 唤醒时**：作为版本管理器运行，负责下载、解压和环境配置。
- **作为 `godot` 唤醒时**：完全跳过所有参数解析，化身为一个 **100% 透明代理**。它会在瞬间定位真实引擎，并通过 `os/exec` 完美移交标准输入输出流和 `Ctrl+C` 等中断信号。

### 2. 动态目录嗅探 (Tree Walk)
当垫片启动时，它会从当前执行命令的目录 (CWD) 开始，逐级向父目录查找 `.gduprc` 配置文件。这使得您可以在不同的项目文件夹下，无感地启动不同版本的 Godot 引擎。如果未找到，将默认回退到家目录的全局配置。

### 3. 环境变量 (GDUP_CACHE_PATH)
GDUP 默认将数十 GB 的引擎本体解压在 `~/.gdup/versions` 中。对于 C 盘空间紧张的 Windows 用户，可自行在系统中添加 `GDUP_CACHE_PATH` 环境变量，程序会自动将引擎安装到该路径下的 `versions` 目录中。

---

## ⚙️ 源码编译

我们提供了跨平台的自动编译与安装脚本，运行后会自动将最新的二进制程序安装到 `~/.gdup/bin`：

**Windows 开发环境 (PowerShell)**
```powershell
.\scripts\build_windows.ps1
```

**macOS / Linux 开发环境**
```bash
./scripts/build.sh
```

---

## 📦 自动化发布 (CI/CD)

项目已集成 GitHub Actions 流水线。如需发布新版本，只需打上以 `v` 开头的 Tag 并推送到远端：

```bash
git tag v1.0.1
git push origin v1.0.1
```

GitHub 的云端构建机将自动进行交叉编译，并将 Windows、macOS (Intel/M系列)、Linux 平台下的 5 个免安装二进制文件发布到 Releases 页面，供终端用户的 Curl 脚本拉取。
