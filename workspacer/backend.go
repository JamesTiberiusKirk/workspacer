package workspacer

import (
	"os"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	gotmux "github.com/jubnzv/go-tmux"
)

// SessionBackend is the multiplexer surface workspacer needs. Orchestration
// (start-or-switch, prefix-kill, sister repos, spec assembly) lives above this;
// a backend only realizes the primitives. tmux and gtmux each implement it.
type SessionBackend interface {
	HasSession(name string) bool
	ListSessions() ([]string, error)
	KillSession(name string) error
	// CreateSession builds a DETACHED session from spec (windows/panes with
	// names, layouts, start-dirs, commands, sizes). Callers check HasSession first.
	CreateSession(spec SessionSpec) error
	// Attach connects the current terminal to name: switch-client if we're already
	// inside a session of this backend, else attach.
	Attach(name string) error
}

// SessionSpec is a backend-agnostic description of a session to build.
type SessionSpec struct {
	Name    string
	Path    string // session root; default start-dir for its windows
	Windows []WindowSpec
}

type WindowSpec struct {
	Name   string
	Layout string // tmux layout name: even-horizontal, main-vertical, tiled, ...
	Path   string // start-dir; "" falls back to SessionSpec.Path
	Panes  []PaneSpec
}

type PaneSpec struct {
	Command string // run in the pane; "" = bare shell
	Size    int    // width percent, 0 = layout default
}

// GetBackend picks the multiplexer from the WORKSPACER_MUX env var (set it in
// your profile next to the aliases). Defaults to tmux when unset/unrecognized.
func GetBackend() SessionBackend {
	if os.Getenv("WORKSPACER_MUX") == string(config.MuxGtmux) {
		return newGtmuxBackend()
	}
	return newTmuxBackend()
}

// CurrentSessionName reports the session this process is running inside, probing
// gtmux ($GTMUX = sock,pid,session) first, then tmux. Backend-independent: the
// middleware needs it before a workspace (hence a backend) is known.
func CurrentSessionName() (string, bool) {
	if g := os.Getenv("GTMUX"); g != "" {
		parts := strings.Split(g, ",")
		if len(parts) >= 3 && parts[2] != "" {
			return parts[2], true
		}
	}
	if os.Getenv("TMUX") != "" {
		if name, err := gotmux.GetAttachedSessionName(); err == nil && name != "" {
			return name, true
		}
	}
	return "", false
}
