# Workspacer

> A powerful tmux session manager for organizing projects across multiple workspaces

[![Go Version](https://img.shields.io/github/go-mod/go-version/JamesTiberiusKirk/workspacer)](https://golang.org)
[![License](https://img.shields.io/github/license/JamesTiberiusKirk/workspacer)](LICENSE)

Workspacer helps you manage multiple development environments (workspaces) with tmux, GitHub integration, and smart project organization. Perfect for juggling multiple repos in one workspace with a lot of context switching or juggling multiple workspaces (if say you use the same machine for personal and work projects).

## 🎨 Philosophy

Workspacer follows a **suckless-inspired design philosophy**:
> were not ready for this quite yet but can definitely experiment

**Two ways to use Workspacer:**

1. **Fork it** - Clone the repo, modify the commands in `cmd/workspacer/main.go`, make it yours
2. **Import it** - Use the repo as a go package you import into your own project, copy `cmd/workspacer/main.go` and go get this package.

## ✨ Features

- 🎯 **Workspace Management** - Separate work, personal, and client projects
- 🚀 **Tmux Integration** - Smart session management with custom layouts
- 🔄 **GitHub Integration** - Search, clone, and track repositories
- 📊 **Usage Tracking** - See which projects you work on most
- 🎨 **TUI Pickers** - Interactive project and session selection
- 🔧 **Custom Layouts** - Define tmux session presets for different project types

## 📦 Installation

### Using the Default Implementation

```bash
git clone https://github.com/JamesTiberiusKirk/workspacer.git
cd workspacer
make install
```

### Forking and Customizing

```bash
# Fork the repo on GitHub, then:
git clone https://github.com/YOUR_USERNAME/workspacer.git
cd workspacer

# Modify cmd/workspacer/main.go to add your custom commands
# Edit the ConfigMap to define your own workflows

make install  # Install your customized version
```

### Using as a Go Package

```bash
# Create your own project
mkdir my-workspace-tool
cd my-workspace-tool
go mod init github.com/yourname/my-workspace-tool

# Get workspacer as a dependency
go get github.com/JamesTiberiusKirk/workspacer

# Copy the main.go as your starting point
curl -o main.go https://raw.githubusercontent.com/JamesTiberiusKirk/workspacer/master/cmd/workspacer/main.go

# Customize main.go with your own commands
# Build your tool
go build -o my-tool .
```

> **Note:** Versioned releases coming soon!

### Shell Completion

```bash
# Install completion (auto-detects your shell)
workspacer completion install

# Or manually for zsh
workspacer completion zsh > ~/.zsh/completion/_workspacer

# Add to ~/.zshrc
fpath=(~/.zsh/completion $fpath)

# Reload shell
exec zsh
```

## 🚀 Quick Start

### 1. Create Your First Config

```bash
workspacer config new
```

This creates `~/.config/workspacer/workspaces.json`

### 2. Edit Your Config

```json
{
  "default_workspace": "personal",
  "workspaces": {
    "personal": {
      "name": "Personal Projects",
      "prefix": "personal",
      "path": "~/Projects/personal",
      "org_github": "your-username",
      "is_org": false,
      "enable_cache": true,
      "enable_usage_tracking": true
    },
    "work": {
      "name": "Work Projects",
      "prefix": "work",
      "path": "~/Projects/work",
      "org_github": "company-org",
      "is_org": true,
      "github_backend": "api",
      "enable_remote_repos": true
    }
  }
}
```

### 3. Start Using Workspacer

```bash
# Open project picker for default workspace
workspacer

# Or specify workspace
workspacer -W work

# Open a specific project
workspacer -W personal my-project
```

## 📖 Usage Guide

### Core Commands

#### Project Management

```bash
# Open interactive project picker
workspacer -W personal

# Open specific project
workspacer -W work api-service

# Create new project
workspacer -W personal new my-app

# Create new project with GitHub repo
workspacer -W personal new my-app --gh --private

# Search GitHub for projects
workspacer -W work search "microservice"
```

#### Session Management

```bash
# List open sessions in workspace
workspacer -W personal list

# Open session picker
workspacer -W work open

# Close all sessions in workspace
workspacer -W personal close-all
```

#### Configuration

```bash
# Create new config file
workspacer config new

# Show loaded config and env file paths
workspacer config list
```

#### Cache Management

```bash
# Show cache statistics
workspacer -W personal cache status

# Clear cache for workspace
workspacer -W personal cache clear
```

#### Usage Statistics

```bash
# Show project access statistics
workspacer -W personal stats
```

### Global Flags

```bash
-W, --workspace <name>    # Specify workspace
-D, --debug              # Enable debug mode
-h, --help               # Show help
```

### Environment Files

Workspacer loads environment variables from `.workspace.env` files:

**Priority:**
1. `{workspace_path}/.workspace.env` - Workspace-specific
2. `~/.workspace.env` - Global fallback

**Example `.workspace.env`:**
```bash
GITHUB_TOKEN=ghp_your_token_here
GIT_AUTHOR_NAME=Your Name
GIT_AUTHOR_EMAIL=you@example.com
```

## 🔧 Configuration Reference

### Workspace Config

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name for the workspace |
| `prefix` | string | Tmux session prefix (e.g., "work-project") |
| `path` | string | Path to workspace directory |
| `org_github` | string | GitHub username or organization |
| `is_org` | bool | Whether GitHub account is an organization |
| `github_backend` | string | `"api"` or `"cli"` for GitHub integration |
| `enable_cache` | bool | Enable project list caching |
| `enable_usage_tracking` | bool | Track project access statistics |
| `enable_remote_repos` | bool | Include remote GitHub repos in listings |
| `enable_git_info` | bool | Show git branch/status in listings |
| `recent_access_window` | int | Number of recent accesses to track (default: 50) |

### Session Presets

Define custom tmux layouts for different project types:

```json
{
  "session_presets": {
    "default": {
      "screens": [
        {
          "name": "editor",
          "panes": [
            { "command": "nvim" }
          ]
        },
        {
          "name": "server",
          "panes": [
            { "command": "npm run dev" }
          ]
        }
      ]
    }
  }
}
```

## 🎯 Tmux Integration

### Keybindings

Add to your `~/.tmux.conf`:

```bash
# Prompt to run workspacer with arguments
bind-key P command-prompt -p "Workspacer args:" \
  "new-window 'workspacer -W=current %1 %2 %3'"

# Filter session tree to current workspace
bind-key s run-shell \
  "tmux choose-tree -Zs -f \"$(workspacer -W=current get-tmux-workspace-filter)\""
```

### Usage in Tmux

```bash
# Within tmux, use current workspace
workspacer -W=current open
```

## 🔍 Advanced Features

### Tenant Repositories

Support for service/tenant repository patterns:

```json
{
  "workspaces": {
    "saas": {
      "enable_tenant_repos": true,
      "tenant_repo_prefix": "tenant-"
    }
  }
}
```

### GitHub Actions Integration

```bash
# Check CI status for current project
workspacer -W=current actions
```

### Custom Aliases

Add to your `~/.zshrc`:

```bash
# Define aliases
alias ws="workspacer -W=personal"
alias work="workspacer -W=work"
alias side="workspacer -W=sideprojects"

# Enable completion for aliases
compdef ws=workspacer
compdef work=workspacer
compdef side=workspacer
```

**For bash**, use wrapper functions instead:

```bash
ws() { workspacer -W=personal "$@"; }
work() { workspacer -W=work "$@"; }
side() { workspacer -W=sideprojects "$@"; }

complete -F _workspacer ws
complete -F _workspacer work
complete -F _workspacer side
```

Now tab completion works:
```bash
ws config <TAB>      # Shows: new, list
work cache <TAB>     # Shows: clear, status
```

## 🏗️ Development & Customization

### Project Structure

```
workspacer/
├── cli/          # CLI framework
├── cmd/          # Application entry point - CUSTOMIZE THIS
├── commands/     # Command implementations - ADD YOUR COMMANDS HERE
├── config/       # Configuration management
├── state/        # Application state
├── util/         # Utility functions
└── workspacer/   # Core workspace management logic
```

### Building Your Own Version

**1. Fork the repository**

**2. Add your custom commands** in `cmd/workspacer/main.go`:

```go
var ConfigMap cli.ConfigMapType = cli.ConfigMapType{
    "my-command": &cli.Command{
        Description: "My custom workflow",
        Flags: []cli.Flag{
            {Name: "option", Type: "string", Description: "My option"},
        },
        Runner: myCustomCommand,
    },
    // ... keep or remove default commands
}
```

**3. Implement your command:**

```go
func myCustomCommand(ctx cli.ConfigMapCtx) {
    option := ctx.GetStringFlag("option")
    // Your custom logic here
}
```

**4. Build and install:**

```bash
make install
```

**5. Completion auto-updates!** The completion script automatically includes your new commands.

### Using Workspacer Functionality in Your Own Tool

You can import workspacer's core packages to use its workspace management features:

**Core Packages:**
- `github.com/JamesTiberiusKirk/workspacer/workspacer` - Workspace management, tmux sessions, GitHub integration
- `github.com/JamesTiberiusKirk/workspacer/config` - Configuration loading and workspace configs
- `github.com/JamesTiberiusKirk/workspacer/commands` - Built-in command implementations you can reuse

**Example - Using Workspacer's Features:**

```go
package main

import (
    "github.com/JamesTiberiusKirk/workspacer/config"
    "github.com/JamesTiberiusKirk/workspacer/workspacer"
)

func main() {
    // Load workspacer config
    cfg, err := config.LoadFromDefaultConfigPath()
    if err != nil {
        panic(err)
    }

    workspaceConfig := cfg.Workspaces["personal"]

    // Use workspacer's functionality
    // Open project picker
    t, choice := workspacer.ChoseProjectFromLocalWorkspace(
        workspaceConfig.Prefix,
        workspaceConfig,
        nil,
    )

    // Start or switch to tmux session
    workspacer.StartOrSwitchToSession(
        workspaceConfig,
        cfg.SessionPresets,
        choice,
    )
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [go-tmux](https://github.com/jubnzv/go-tmux)
- Inspired by [tmuxinator](https://github.com/tmuxinator/tmuxinator) and [tmux-sessionizer](https://github.com/ThePrimeagen/.dotfiles)

## 📬 Support

- 🐛 [Report a bug](https://github.com/JamesTiberiusKirk/workspacer/issues)
- 💡 [Request a feature](https://github.com/JamesTiberiusKirk/workspacer/issues)
- 📖 [Documentation](https://github.com/JamesTiberiusKirk/workspacer/wiki)

---

**Made with ❤️ for developers who live in the terminal**
