package workspacer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/util"
)

func CloneRepo(wc config.WorkspaceConfig, repoName string) error {
	parentFolder, err := util.ExpandTilde(wc.Path)
	if err != nil {
		return fmt.Errorf("failed to expand workspace path: %w", err)
	}

	githubUserOrOrg := wc.GithubOrg
	repoURL := fmt.Sprintf("git@github.com:%s/%s.git", githubUserOrOrg, repoName)
	clonePath := filepath.Join(parentFolder, repoName)

	// Check if destination directory already exists
	if _, err := os.Stat(clonePath); err == nil {
		return fmt.Errorf("repository '%s' already exists at %s", repoName, clonePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if path exists: %w", err)
	}

	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not in PATH: %w", err)
	}

	// Run git clone
	cmd := exec.Command("git", "clone", repoURL, clonePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Cloning %s into %s...\n", repoURL, clonePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Println("Clone complete.")
	return nil
}

func NewProjectAndPush(wc config.WorkspaceConfig, repoName string) error {
	parentFolder, err := util.ExpandTilde(wc.Path)
	if err != nil {
		return fmt.Errorf("failed to expand workspace path: %w", err)
	}

	repoPath := filepath.Join(parentFolder, repoName)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	fmt.Printf("Creating folder %s\n", repoPath)

	// Git init with master
	if _, err := util.ExecCmd("", "git", "-C", repoPath, "init", "-b", "master"); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	fmt.Printf("Initialised git repo\n")

	// Derive go module path
	modulePath := fmt.Sprintf("github.com/%s/%s", wc.GithubOrg, repoName)
	if _, err := util.ExecCmd(repoPath, "go", "mod", "init", modulePath); err != nil {
		return fmt.Errorf("go mod init failed: %w", err)
	}

	fmt.Printf("Initialised go module %s\n", modulePath)

	// Create README
	readmePath := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(readmePath, []byte("# "+repoName+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	fmt.Printf("Created README.md\n")

	if _, err := util.ExecCmd("", "git", "-C", repoPath, "add", "."); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	if _, err := util.ExecCmd("", "git", "-C", repoPath, "commit", "-m", "'Initial commit'"); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Printf("Created initial commit\n")

	sshURL := fmt.Sprintf("git@github.com:%s/%s.git", wc.GithubOrg, repoName)

	if _, err := util.ExecCmd("", "git", "-C", repoPath, "remote", "add", "origin", sshURL); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}
	if _, err := util.ExecCmd("", "git", "-C", repoPath, "push", "-u", "--force", "origin", "master"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	fmt.Println("Repository created and pushed to GitHub:", sshURL)
	return nil
}
