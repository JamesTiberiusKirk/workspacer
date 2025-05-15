package util

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/joho/godotenv"
	gotmux "github.com/jubnzv/go-tmux"
)

const (
	envFileName = ".workspace.env"
)

func LoadEnvFile(wc config.WorkspaceConfig) {
	path := GetWorkspacePath(wc) + "/" + envFileName
	if _, err := os.Stat(path); os.IsNotExist(err) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		if _, err := os.Stat(homeDir + "/" + envFileName); !os.IsNotExist(err) {
			path = homeDir + "/" + envFileName
		} else {
			log.Debug("No .workspace.env file found in workspace or home directory")
			return
		}
	}

	err := godotenv.Load(path)
	if err != nil {
		panic(err)
	}
}

func GetOpenProjectsByWorkspace(wsPrefix string) []string {
	server := new(gotmux.Server)
	sessions, err := server.ListSessions()
	if err != nil {
		log.Error("could not get tmux sessions: %s\n",
			err.Error())
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

func GetProjectPath(wc config.WorkspaceConfig, project string) string {
	projectPath := GetWorkspacePath(wc) + "/" + project

	_, err := os.Stat(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("project path does not exist: %s", projectPath)
			return ""
		}

		log.Error("could not get stat on path: %s", err.Error())
		return ""
	}

	return projectPath
}

func DoesBranchExist(wc config.WorkspaceConfig, project, branch string) bool {
	projectPath := GetProjectPath(wc, project)

	gitOut, err := ExecCmd("", "git", "-C", projectPath, "remote", "show", "origin")
	if err != nil {
		return false
	}

	at := strings.Index(gitOut, "HEAD branch")
	if at == -1 {
		return false
	}

	if !strings.Contains(gitOut, branch) {
		return false
	}

	return true
}

func GetGitMainBranch(wc config.WorkspaceConfig, project string) string {
	projectPath := GetProjectPath(wc, project)

	gitOut, err := ExecCmd("", "git", "-C", projectPath, "remote", "show", "origin")
	if err != nil {
		log.Error("could not exec git: %s", err.Error())
		return ""
	}

	at := strings.Index(gitOut, "HEAD branch")
	if at == -1 {
		log.Error("could not find HEAD branch in git output")
	}

	branch := gitOut[at:]
	branch = strings.TrimPrefix(branch, "HEAD branch: ")
	branch = strings.Split(branch, " ")[0]
	branch = strings.TrimSpace(branch)

	return branch
}

func GetProjectCurrentBranch(wc config.WorkspaceConfig, project string) string {
	projectPath := GetProjectPath(wc, project)

	branch, err := ExecCmd("", "git", "-C", projectPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		log.Error("could not exec git: %s", err.Error())
		return ""
	}

	return branch
}

func ExecCmd(path, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if path != "" {
		cmd.Dir = path
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func ExpandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[1:]), nil
	}
	return path, nil
}
