package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	gotmux "github.com/jubnzv/go-tmux"
)

func GetOpenProjectsByWorkspace(wsPrefix string) []string {
	server := new(gotmux.Server)
	sessions, err := server.ListSessions()
	if err != nil {
		log.Error("could not get tmux sessions: %s", err.Error())
		return []string{}
	}

	openProjects := []string{}
	for _, s := range sessions {
		if !strings.HasPrefix(s.Name, wsPrefix) {
			continue
		}
		openProjects = append(openProjects, strings.TrimPrefix(s.Name, wsPrefix+"-"))
	}

	return openProjects
}

func GetWorkspacePath(wc config.WorkspaceConfig) string {
	if strings.HasPrefix(wc.Path, "~/") {
		dirname, _ := os.UserHomeDir()
		wc.Path = filepath.Join(dirname, wc.Path[2:])
	}

	return wc.Path
}

func HasGitSubfolder(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func DoesProjectExist(wc config.WorkspaceConfig, project string) bool {
	path := filepath.Join(GetWorkspacePath(wc), project)
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !s.IsDir() {
		return false
	}

	return true
}
