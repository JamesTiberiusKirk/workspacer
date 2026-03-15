package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
	Name          string `yaml:"name"`
	SubPath       string `yaml:"sub_path,omitempty"`
	SessionPreset string `yaml:"session_preset,omitempty"`
}

type GithubBackend string

const (
	GithubBackendAPI GithubBackend = "api"
	GithubBackendCLI GithubBackend = "cli"
)

type WorkspaceConfig struct {
	Name                string          `yaml:"name"`
	Prefix              string          `yaml:"prefix"`
	Path                string          `yaml:"path"`
	GithubOrg           string          `yaml:"org_github,omitempty"`
	IsOrg               bool            `yaml:"is_org,omitempty"`
	Projects            []ProjectConfig `yaml:"projects,omitempty"`
	SessionPreset       string          `yaml:"session_preset,omitempty"`
	Session             *SessionConfig  `yaml:"session_config,omitempty"`
	EnableTenantRepos   bool            `yaml:"enable_tenant_repos,omitempty"`
	TenantRepoPrefix    string          `yaml:"tenant_repo_prefix,omitempty"`
	ActiveProjectsFirst bool            `yaml:"active_projects_first,omitempty"`
	EnableGitInfo       bool            `yaml:"enable_git_info,omitempty"`
	EnableRemoteRepos   bool            `yaml:"enable_remote_repos,omitempty"`
	GithubBackend       GithubBackend   `yaml:"github_backend,omitempty"` // "api" or "cli", defaults to "api"
	EnableCache         bool            `yaml:"enable_cache,omitempty"`
	EnableUsageTracking bool            `yaml:"enable_usage_tracking,omitempty"`
	RecentAccessWindow  int             `yaml:"recent_access_window,omitempty"` // Default: 50
}

type PanesConfig struct {
	Command     string      `yaml:"command,omitempty"`
	Orientation Orientation `yaml:"orientation,omitempty"`
	Size        int         `yaml:"size,omitempty"`
	Path        string      `yaml:"path,omitempty"`
}

type WindowConfig struct {
	Panes  []PanesConfig `yaml:"panes,omitempty"`
	Name   string        `yaml:"name,omitempty"`
	Layout string        `yaml:"layout,omitempty"`
	Path   string        `yaml:"path,omitempty"`
}

type SessionConfig struct {
	Windows []WindowConfig `yaml:"screens,omitempty"`
	Path    string         `yaml:"path,omitempty"`
}

func (c *SessionConfig) ListPanes() []PanesConfig {
	panes := []PanesConfig{}
	for _, w := range c.Windows {
		panes = append(panes, w.Panes...)
	}
	return panes
}

type GlobalUserConfig struct {
	DefaultWorkspace string                     `yaml:"default_workspace,omitempty"`
	Workspaces       map[string]WorkspaceConfig `yaml:"workspaces,omitempty"`
	SessionPresets   map[string]SessionConfig   `yaml:"session_presets,omitempty"`
	GitPath          string                     `yaml:"git_path,omitempty"`
	GithubPath       string                     `yaml:"github_path,omitempty"`
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
	defaultConfigFile = "workspaces.yaml"
)

// GetDefaultConfigPath returns the full path to the default config file
func GetDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	return filepath.Join(homeDir, defaultConfigPath, defaultConfigFile), nil
}

// WriteDefaultConfig creates the default config directory and writes a blank config to file
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

	// Write a blank config to file
	b, err := yaml.Marshal(BlankConfig)
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
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		fmt.Printf("Error unmarshaling config: %s\n", err.Error())
		return nil, err
	}

	return &conf, nil
}

func LoadFromDefaultConfigPath() (*GlobalUserConfig, error) {
	fmt.Println("Loading global config")

	configPath, err := GetDefaultConfigPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return LoadGlobalConfig(configPath)
}

func WriteConfigToFile(conf GlobalUserConfig, path string) error {
	fmt.Printf("Writting config to file: %s\n", path)

	b, err := yaml.Marshal(conf)
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
