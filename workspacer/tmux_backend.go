package workspacer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	gotmux "github.com/jubnzv/go-tmux"
)

// tmuxBackend drives real tmux via go-tmux (declarative Configuration.Apply)
// plus a few direct tmux commands. This is the logic that used to live inline in
// tmux.go's StartOrSwitchTo* functions, moved behind SessionBackend unchanged.
type tmuxBackend struct{}

func newTmuxBackend() *tmuxBackend { return &tmuxBackend{} }

func (b *tmuxBackend) HasSession(name string) bool {
	// go-tmux's HasSession errors on an empty server, so list + scan instead
	// (matches the len(sessions)>0 guard the old code used).
	server := new(gotmux.Server)
	sessions, err := server.ListSessions()
	if err != nil || len(sessions) == 0 {
		return false
	}
	for _, s := range sessions {
		if s.Name == name {
			return true
		}
	}
	return false
}

func (b *tmuxBackend) ListSessions() ([]string, error) {
	server := new(gotmux.Server)
	sessions, err := server.ListSessions()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(sessions))
	for i, s := range sessions {
		names[i] = s.Name
	}
	return names, nil
}

func (b *tmuxBackend) KillSession(name string) error {
	server := new(gotmux.Server)
	return server.KillSession(name)
}

func (b *tmuxBackend) CreateSession(spec SessionSpec) error {
	server := new(gotmux.Server)

	windows := make([]gotmux.Window, 0, len(spec.Windows))
	for i, w := range spec.Windows {
		dir := w.Path
		if dir == "" {
			dir = spec.Path
		}
		windows = append(windows, gotmux.Window{
			Id:             i + 1,
			Name:           w.Name,
			Layout:         w.Layout,
			Panes:          make([]gotmux.Pane, len(w.Panes)),
			StartDirectory: dir,
		})
	}

	session := gotmux.NewSession(0, spec.Name, spec.Path, windows)
	server.AddSession(*session)
	conf := gotmux.Configuration{Server: server, Sessions: []*gotmux.Session{session}, ActiveSession: nil}
	if err := conf.Apply(); err != nil {
		return fmt.Errorf("apply tmux configuration: %w", err)
	}

	// Run each pane's command and size it, window by window.
	sessWindows, err := session.ListWindows()
	if err != nil {
		return fmt.Errorf("list windows: %w", err)
	}
	for wi, w := range spec.Windows {
		if wi >= len(sessWindows) {
			break
		}
		wPanes, err := sessWindows[wi].ListPanes()
		if err != nil {
			continue
		}
		for pi, p := range w.Panes {
			if pi >= len(wPanes) {
				break
			}
			if p.Size > 0 && p.Size < 100 {
				b.paneSize(p.Size)
			}
			if p.Command != "" {
				wPanes[pi].RunCommand(p.Command)
			}
		}
	}

	// Focus the first window/pane (tmux leaves the last-built one active).
	if len(sessWindows) > 0 {
		sessWindows[0].Select()
		if p, err := sessWindows[0].ListPanes(); err == nil && len(p) > 0 {
			p[0].Select()
		}
	}
	return nil
}

func (b *tmuxBackend) Attach(name string) error {
	var cmd *exec.Cmd
	if os.Getenv("TMUX") != "" {
		cmd = exec.Command("tmux", "switch-client", "-t="+name)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t="+name)
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// paneSize resizes the pane to the right to size% width (the old paneSize()
// helper, unchanged — it targets {right}, assuming the common two-pane split).
func (b *tmuxBackend) paneSize(size int) {
	out, errStr, errExec := tmuxCmd([]string{"resizep", "-t{right}", "-x " + fmt.Sprint(size) + "%"})
	if errExec != nil {
		fmt.Print(errExec.Error())
	}
	if errStr != "" {
		fmt.Print(errStr)
	}
	fmt.Print(out)
}

// tmuxCmd runs a raw tmux command (the old RunCmd helper).
func tmuxCmd(args []string) (string, string, error) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(tmux, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}
