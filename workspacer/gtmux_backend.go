package workspacer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// gtmuxBackend drives github.com/FyrmForge/gtmux via its CLI. gtmux builds a
// session imperatively (new -d, then run <s> new-window/split-window/... against
// the active window/pane), unlike tmux's declarative go-tmux Apply.
type gtmuxBackend struct{ bin string }

func newGtmuxBackend() *gtmuxBackend {
	bin := os.Getenv("GTMUX_BIN")
	if bin == "" {
		bin = "gtmux"
	}
	return &gtmuxBackend{bin: bin}
}

func (b *gtmuxBackend) HasSession(name string) bool {
	// `gtmux has <name>` exits 0 iff it exists.
	return exec.Command(b.bin, "has", name).Run() == nil
}

func (b *gtmuxBackend) ListSessions() ([]string, error) {
	out, err := exec.Command(b.bin, "list").Output()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "no sessions" {
			continue
		}
		// each line is "<name>: <n> windows".
		if i := strings.IndexByte(line, ':'); i > 0 {
			names = append(names, line[:i])
		}
	}
	return names, nil
}

func (b *gtmuxBackend) KillSession(name string) error {
	return exec.Command(b.bin, "kill-session", name).Run()
}

func (b *gtmuxBackend) CreateSession(spec SessionSpec) error {
	// Detached create; the session's first window/pane inherits this cwd.
	newCmd := exec.Command(b.bin, "new", "-d", spec.Name)
	newCmd.Dir = spec.Path
	if err := newCmd.Run(); err != nil {
		return fmt.Errorf("gtmux new -d %s: %w", spec.Name, err)
	}

	for i, w := range spec.Windows {
		wdir := w.Path
		if wdir == "" {
			wdir = spec.Path
		}
		if i == 0 {
			if w.Name != "" {
				b.run(spec.Name, "rename-window", w.Name)
			}
			// window[0]/pane[0] already exists, rooted at spec.Path.
		} else {
			args := []string{"new-window"}
			if w.Name != "" {
				args = append(args, "-n", w.Name)
			}
			args = append(args, "-c", wdir)
			b.run(spec.Name, args...)
		}

		for pi, p := range w.Panes {
			if pi > 0 {
				b.run(spec.Name, "split-window", "-c", wdir)
			}
			if p.Command != "" {
				// -l = literal text (avoids key-name lookup), then Enter.
				b.run(spec.Name, "send-keys", "-l", p.Command)
				b.run(spec.Name, "send-keys", "Enter")
			}
			if p.Size > 0 && p.Size < 100 {
				b.run(spec.Name, "resize-pane", "-x", fmt.Sprintf("%d%%", p.Size))
			}
		}
		if w.Layout != "" {
			b.run(spec.Name, "select-layout", w.Layout)
		}
	}

	// Focus the first window (base-index-independent).
	b.run(spec.Name, "select-window", "{start}")
	return nil
}

func (b *gtmuxBackend) Attach(name string) error {
	// ponytail: always attach — gtmux has no client-side switch-client CLI, and
	// workspacer is normally launched from outside the multiplexer. Nested attach
	// (already inside gtmux) opens a second client, acceptable for the POC.
	cmd := exec.Command(b.bin, "attach", name)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// run executes `gtmux run <session> <args...>` best-effort, printing any error
// (mirrors the tmux backend's tolerance of per-step failures during build).
func (b *gtmuxBackend) run(session string, args ...string) {
	full := append([]string{"run", session}, args...)
	if out, err := exec.Command(b.bin, full...).CombinedOutput(); err != nil {
		fmt.Printf("gtmux run %s %s: %v %s\n", session, strings.Join(args, " "), err, out)
	}
}
