<p align="right">
  <strong>English</strong> | <a href="README_zh.md">简体中文</a>
</p>

# GDUP - Godot Version Manager

GDUP is a lightweight, cross-platform Godot Engine version manager natively written in Go. It aims to help developers efficiently install, manage, and seamlessly switch between multiple Godot engine versions on a single machine for parallel multi-project development.

## ✨ Features

- **Project-Level Version Isolation**: Automatic configuration recognition based on directory structure. By creating a `.gduprc` file in the project root, the system automatically matches and launches the corresponding engine version for different projects.
- **Transparent Proxy Architecture (Shim)**: Provides an optional global command hijacking feature. Once enabled, developers can directly execute the standard `godot` command in the terminal. GDUP will securely proxy it to the target version engine with zero latency, without changing existing development habits.
- **Zero External Dependencies**: Distributed as a single statically compiled binary file, requiring no pre-installed runtime environments (such as Python, Node.js, etc.).
- **Centralized Storage**: All downloaded engine versions are centrally maintained in the `~/.gdup/versions` directory. Supports customizing the storage path via the `GDUP_CACHE_PATH` environment variable to prevent disk space bloat.
- **Native Cross-Platform Support**: Provides native binary support for Windows, macOS (Intel/Apple Silicon), and Linux.

---

## 🚀 Installation

The system will automatically download the latest binary file suitable for your environment and install it uniformly into `~/.gdup/bin` in your user directory.

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.ps1 | iex
```

### macOS / Linux (Terminal)
```bash
curl -fsSL https://raw.githubusercontent.com/yinfall/gdup/main/scripts/install.sh | bash
```

> **Note**: After installation, please ensure that `~/.gdup/bin` (or `C:\Users\<YourUsername>\.gdup\bin` on Windows) is added to your operating system's global environment variable `PATH`.

---

## 🛠️ Basic Usage

Through the `gdup` main command, you can perform full life-cycle management of engine versions:

```bash
gdup releases         # List available official Godot releases remotely
gdup install 4.3      # Download and install a specific version
gdup use 4.3          # Configure the specific version for the current working directory (generates .gduprc)
gdup list             # List all locally installed versions
gdup godot --editor   # Explicitly launch the Godot engine matching the current project via GDUP
```

### Proxy Configuration (Shim)

If you wish to use the native `godot` command directly (instead of `gdup godot`), you can enable the transparent proxy mode:

```bash
# Install and enable transparent proxy
gdup shim install
```
Once enabled, you can directly execute `godot` in any project directory configured with `.gduprc`, and the system will automatically complete the version routing at the bottom level.

```bash
# Remove transparent proxy and restore system default environment
gdup shim remove
```

---

## 📚 Developer Documentation

For details on GDUP's underlying directory sniffing mechanism, Monolithic Application (Fat Binary) architecture design, environment variable configuration instructions, and how to contribute, please refer to the [Developer and Architecture Documentation](docs/development.md).

## 📄 License
This project is licensed under the [MIT License](LICENSE).
