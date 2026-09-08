package workspacer

import "testing"

func TestSanitizeNameByBackend(t *testing.T) {
	if got := newTmuxBackend().SanitizeName("go.mod.thing"); got != "go_mod_thing" {
		t.Errorf("tmux: got %q", got)
	}
	if got := newGtmuxBackend().SanitizeName("go.mod.thing"); got != "go.mod.thing" {
		t.Errorf("gtmux: got %q", got)
	}
}
