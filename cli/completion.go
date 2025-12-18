package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateZshCompletion generates a zsh completion script by introspecting the ConfigMap
func GenerateZshCompletion(configMap ConfigMapType, binaryName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("#compdef %s\n\n", binaryName))
	sb.WriteString(fmt.Sprintf("_%s() {\n", binaryName))
	sb.WriteString("    local -a commands\n")
	sb.WriteString("    commands=(\n")

	// Generate command list from ConfigMap
	for key, cmd := range configMap {
		// Skip special command types and hidden commands
		if key == CommandTypeNoCommand || key == CommandTypeDefault || cmd.Description == "" {
			continue
		}

		// Handle comma-separated aliases
		names := strings.Split(key, ",")
		primaryName := names[0]

		// Escape description for zsh
		desc := strings.ReplaceAll(cmd.Description, "'", "\\'")
		sb.WriteString(fmt.Sprintf("        '%s:%s'\n", primaryName, desc))
	}

	sb.WriteString("    )\n\n")

	// Generate global flags from GlobalFlags
	sb.WriteString("    _arguments -C \\\n")
	for _, flag := range GlobalFlags {
		flagDesc := strings.ReplaceAll(flag.Description, "'", "\\'")
		if flag.Short != "" {
			sb.WriteString(fmt.Sprintf("        '(-%s --%s)'{-%s,--%s}'[%s]",
				flag.Short, flag.Name, flag.Short, flag.Name, flagDesc))
		} else {
			sb.WriteString(fmt.Sprintf("        '--%s[%s]", flag.Name, flagDesc))
		}

		// Add value placeholder for non-bool flags
		if flag.Type != "bool" {
			sb.WriteString(fmt.Sprintf(":%s:", flag.Name))
		}
		sb.WriteString("' \\\n")
	}
	sb.WriteString("        '1: :->cmds' \\\n")
	sb.WriteString("        '*::arg:->args'\n\n")

	// Generate case statement for subcommands
	sb.WriteString("    case \"$state\" in\n")
	sb.WriteString("        cmds)\n")
	sb.WriteString("            _describe 'command' commands\n")
	sb.WriteString("            ;;\n")
	sb.WriteString("        args)\n")
	sb.WriteString("            case $words[1] in\n")

	// For each command that might have subcommands or flags
	for key, cmd := range configMap {
		if key == CommandTypeNoCommand || key == CommandTypeDefault || cmd.Description == "" {
			continue
		}

		names := strings.Split(key, ",")
		primaryName := names[0]

		sb.WriteString(fmt.Sprintf("                %s)\n", primaryName))

		// Generate subcommands if they exist
		if len(cmd.Subcommands) > 0 {
			sb.WriteString("                    local -a subcommands\n")
			sb.WriteString("                    subcommands=(\n")
			for subKey, subCmd := range cmd.Subcommands {
				if subCmd.Description == "" {
					continue
				}
				subNames := strings.Split(subKey, ",")
				subPrimaryName := subNames[0]
				subDesc := strings.ReplaceAll(subCmd.Description, "'", "\\'")
				sb.WriteString(fmt.Sprintf("                        '%s:%s'\n", subPrimaryName, subDesc))
			}
			sb.WriteString("                    )\n")
			sb.WriteString("                    _describe 'subcommand' subcommands\n")
			sb.WriteString("                    ;;\n")
			continue
		}

		// Generate flags for this command
		if len(cmd.Flags) > 0 {
			sb.WriteString("                    _arguments \\\n")
			for i, flag := range cmd.Flags {
				flagDesc := strings.ReplaceAll(flag.Description, "'", "\\'")

				// Format flag spec
				var flagSpec string
				if flag.Short != "" {
					// Both short and long form
					flagSpec = fmt.Sprintf("'(-%s --%s)'{-%s,--%s}'[%s]",
						flag.Short, flag.Name, flag.Short, flag.Name, flagDesc)
				} else {
					// Long form only
					flagSpec = fmt.Sprintf("'--%s[%s]", flag.Name, flagDesc)
				}

				// Add value placeholder for non-bool flags
				if flag.Type != "bool" {
					flagSpec += fmt.Sprintf(":%s:", flag.Name)
				}
				flagSpec += "'"

				// Add line continuation except for last flag
				if i < len(cmd.Flags)-1 {
					sb.WriteString(fmt.Sprintf("                        %s \\\n", flagSpec))
				} else {
					sb.WriteString(fmt.Sprintf("                        %s\n", flagSpec))
				}
			}
			sb.WriteString("                    ;;\n")
		} else {
			sb.WriteString("                    ;;\n")
		}
	}

	sb.WriteString("            esac\n")
	sb.WriteString("            ;;\n")
	sb.WriteString("    esac\n")
	sb.WriteString("}\n\n")
	sb.WriteString(fmt.Sprintf("_%s\n", binaryName))

	return sb.String()
}

// GenerateBashCompletion generates a bash completion script (TODO)
func GenerateBashCompletion(configMap ConfigMapType, binaryName string) string {
	return "# Bash completion not yet implemented\n"
}

// installCompletion detects the shell and installs the completion script
func installCompletion(configMap ConfigMapType, binaryName string) {
	// Detect shell from $SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell == "" {
		fmt.Println("Error: Could not detect shell from $SHELL environment variable")
		return
	}

	var completionScript string
	var installPath string

	if strings.Contains(shell, "zsh") {
		// Generate zsh completion
		completionScript = GenerateZshCompletion(configMap, binaryName)

		// Try user directory first (no sudo needed)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %s\n", err)
			return
		}

		completionDir := filepath.Join(homeDir, ".zsh", "completion")
		installPath = filepath.Join(completionDir, "_"+binaryName)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(completionDir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %s\n", completionDir, err)
			return
		}

		// Write completion file
		if err := os.WriteFile(installPath, []byte(completionScript), 0644); err != nil {
			fmt.Printf("Error writing completion file: %s\n", err)
			return
		}

		fmt.Printf("✓ Installed zsh completion to: %s\n", installPath)
		fmt.Println("\nTo enable completion, add this to your ~/.zshrc:")
		fmt.Printf("  fpath=(%s $fpath)\n", completionDir)
		fmt.Println("\nThen reload your shell:")
		fmt.Println("  exec zsh")

	} else if strings.Contains(shell, "bash") {
		fmt.Println("Bash completion not yet implemented")
		fmt.Println("Use: workspacer completion bash > /etc/bash_completion.d/" + binaryName)
	} else {
		fmt.Printf("Unsupported shell: %s\n", shell)
		fmt.Println("Supported shells: zsh, bash")
	}
}

// MakeCompletionCommand creates a completion command that introspects the given ConfigMap
func MakeCompletionCommand(configMap ConfigMapType, binaryName string) *Command {
	return &Command{
		Description: "Generate shell completion scripts",
		Runner: func(ctx ConfigMapCtx) {
			completionSubcommands := ConfigMapType{
				"zsh": {
					Description: "Generate zsh completion script",
					Runner: func(ctx ConfigMapCtx) {
						fmt.Print(GenerateZshCompletion(configMap, binaryName))
					},
				},
				"bash": {
					Description: "Generate bash completion script",
					Runner: func(ctx ConfigMapCtx) {
						fmt.Print(GenerateBashCompletion(configMap, binaryName))
					},
				},
				"install": {
					Description: "Install completion script for current shell",
					Runner: func(ctx ConfigMapCtx) {
						installCompletion(configMap, binaryName)
					},
				},
			}

			HandleSubcommands(ctx, completionSubcommands, "Usage: completion [zsh|bash|install]")
		},
	}
}
