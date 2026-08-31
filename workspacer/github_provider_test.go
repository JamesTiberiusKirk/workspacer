package workspacer

import (
	"testing"

	"github.com/JamesTiberiusKirk/workspacer/config"
)

func TestGetProviderPicksBackend(t *testing.T) {
	if _, ok := GetProvider(config.WorkspaceConfig{GithubBackend: config.GithubBackendCLI}).(*CLIProvider); !ok {
		t.Fatal("cli backend should give CLIProvider")
	}
	if _, ok := GetProvider(config.WorkspaceConfig{}).(*APIProvider); !ok {
		t.Fatal("empty backend should default to APIProvider")
	}
}

func TestGithubTokenPrefersEnv(t *testing.T) {
	t.Setenv("GITHUB_AUTH", "tok123")
	got, err := githubToken()
	if err != nil || got != "tok123" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestGithubTokenErrorsWithoutAuth(t *testing.T) {
	t.Setenv("GITHUB_AUTH", "")
	t.Setenv("PATH", t.TempDir()) // no gh CLI available
	if _, err := githubToken(); err == nil {
		t.Fatal("expected an error, not a silent anonymous client")
	}
}
