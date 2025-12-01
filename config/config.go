package config

import (
	"encoding/json"
	"fmt"
	"os"
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
	SubPath       string `json:"sub_path"`
	SessionPreset string
}

type WorkspaceConfig struct {
	Name              string          `json:"name"`
	Prefix            string          `json:"prefix"`
	Path              string          `json:"path"`
	GithubOrg         string          `json:"org_github"`
	IsOrg             bool            `json:"is_org"`
	Projects          []ProjectConfig `json:"projects"`
	SessionPreset     string          `json:"session_preset"`
	Session           SessionConfig   `json:"session_config"`
	EnableTenantRepos bool            `json:"enable_tenant_repos"`
	TenantRepoPrefix  string          `json:"tenant_repo_prefix"`
	ActiveProjectsFirst bool          `json:"active_projects_first"`
}

type PanesConfig struct {
	Command     string      `json:"command"`
	Orientation Orientation `json:"orientation"`
	Size        int         `json:"size"`
	Path        string
}

type WindowConfig struct {
	Panes  []PanesConfig `json:"panes"`
	Name   string        `json:"name"`
	Layout string        `json:"layout"`
	Path   string
}

type SessionConfig struct {
	Windows []WindowConfig `json:"screens"`
	Path    string
}

func (c *SessionConfig) ListPanes() []PanesConfig {
	panes := []PanesConfig{}
	for _, w := range c.Windows {
		panes = append(panes, w.Panes...)
	}
	return panes
}

type GlobalUserConfig struct {
	DefaultWorkspace string                     `json:"default_workspace"`
	Workspaces       map[string]WorkspaceConfig `json:"workspaces"`
	SessionPresets   map[string]SessionConfig   `json:"session_presets"`
	GitPath          string                     `json:"git_path"`
	GithubPath       string                     `json:"github_path"`
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
