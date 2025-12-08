package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Orientation string

const (
	OrientationHorizontal Orientation = "horizontal"
	OrientationVertical               = "vertical"
)

func (o Orientation) IsValid() bool {
	switch o {
	case OrientationHorizontal, OrientationVertical:
		return true
	}

	return false
}

type ProjectConfig struct {
	Name          string `json:"name"`
	SubPath       string `json:"sub_path,omitempty"`
	SessionPreset string `json:"session_preset,omitempty"`
}

type GithubBackend string

const (
	GithubBackendAPI GithubBackend = "api"
	GithubBackendCLI GithubBackend = "cli"
)

type WorkspaceConfig struct {
	Name                 string           `json:"name"`
	Prefix               string           `json:"prefix"`
	Path                 string           `json:"path"`
	GithubOrg            string           `json:"org_github,omitempty"`
	IsOrg                bool             `json:"is_org,omitempty"`
	Projects             []ProjectConfig  `json:"projects,omitempty"`
	SessionPreset        string           `json:"session_preset,omitempty"`
	Session              *SessionConfig   `json:"session_config,omitempty"`
	EnableTenantRepos    bool             `json:"enable_tenant_repos,omitempty"`
	TenantRepoPrefix     string           `json:"tenant_repo_prefix,omitempty"`
	ActiveProjectsFirst  bool             `json:"active_projects_first,omitempty"`
	EnableGitInfo        bool             `json:"enable_git_info,omitempty"`
	EnableRemoteRepos    bool             `json:"enable_remote_repos,omitempty"`
	GithubBackend        GithubBackend    `json:"github_backend,omitempty"` // "api" or "cli", defaults to "api"
	EnableCache          bool             `json:"enable_cache,omitempty"`
	EnableUsageTracking  bool             `json:"enable_usage_tracking,omitempty"`
	RecentAccessWindow   int              `json:"recent_access_window,omitempty"` // Default: 50
}

type PanesConfig struct {
	Command     string      `json:"command,omitempty"`
	Orientation Orientation `json:"orientation,omitempty"`
	Size        int         `json:"size,omitempty"`
	Path        string      `json:"path,omitempty"`
}

type WindowConfig struct {
	Panes  []PanesConfig `json:"panes,omitempty"`
	Name   string        `json:"name,omitempty"`
	Layout string        `json:"layout,omitempty"`
	Path   string        `json:"path,omitempty"`
}

type SessionConfig struct {
	Windows []WindowConfig `json:"screens,omitempty"`
	Path    string         `json:"path,omitempty"`
}

func (c *SessionConfig) ListPanes() []PanesConfig {
	panes := []PanesConfig{}
	for _, w := range c.Windows {
		panes = append(panes, w.Panes...)
	}
	return panes
}

type GlobalUserConfig struct {
	DefaultWorkspace string                     `json:"default_workspace,omitempty"`
	Workspaces       map[string]WorkspaceConfig `json:"workspaces,omitempty"`
	SessionPresets   map[string]SessionConfig   `json:"session_presets,omitempty"`
	GitPath          string                     `json:"git_path,omitempty"`
	GithubPath       string                     `json:"github_path,omitempty"`
}

func (c *GlobalUserConfig) GetDefaultWorkspaceConf() (WorkspaceConfig, error) {
	wc, ok := c.Workspaces[c.DefaultWorkspace]
	if !ok {
		return WorkspaceConfig{}, fmt.Errorf("Default workspace not found")
	}

	return wc, nil
}

func (c *GlobalUserConfig) IsValid() bool {
	// TODO:
	//	- Need to validate the different paths
	//	- Need to validate session preset names
	return false
}

const (
	defaultConfigPath = ".config/workspacer/"
	defaultConfigFile = "workspaces.json"
)

// GetDefaultConfigPath returns the full path to the default config file
func GetDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	return filepath.Join(homeDir, defaultConfigPath, defaultConfigFile), nil
}

// WriteDefaultConfig creates the default config directory and writes the DefaultGlobalConfig to file
func WriteDefaultConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, defaultConfigPath)
	configPath := filepath.Join(configDir, defaultConfigFile)

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	// Write the DefaultGlobalConfig to file
	b, err := json.MarshalIndent(DefaultGlobalConfig, "", "\t")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, b, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

func LoadGlobalConfig(path string) (*GlobalUserConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err.Error())
		return nil, err
	}

	conf := GlobalUserConfig{}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		fmt.Printf("Error unmarshaling bytes: %s\n", err.Error())
		return nil, err
	}

	return &conf, nil
}

func LoadFromDefaultConfigPath() (*GlobalUserConfig, error) {
	fmt.Println("Loading global config")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home dir: %s\n", err.Error())
		return nil, err
	}

	_, err = os.Stat(homeDir + defaultConfigPath)
	if err != nil {
		if !os.IsExist(err) {
			return nil, nil
		}
		fmt.Printf("Error getting stat on default config path: %s\n", err.Error())
		return nil, err
	}

	configPath := homeDir + defaultConfigPath + defaultConfigFile

	_, err = os.Stat(configPath)
	if err != nil {
		if os.IsExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return LoadGlobalConfig(configPath)
}

func WriteConfigToFile(conf GlobalUserConfig, path string) error {
	fmt.Printf("Writting config to file: %s\n", path)

	b, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		fmt.Printf("Error marshlling conf: %s\n", err.Error())
		return err
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %s err: %s\n", path, err.Error())
		return err
	}

	_, err = file.Write(b)
	if err != nil {
		fmt.Printf("Error writting to file: %s err: %s\n", path, err.Error())
		return err
	}

	return nil
}
