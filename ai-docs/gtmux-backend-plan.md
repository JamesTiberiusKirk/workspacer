# Pluggable session backend (tmux / gtmux)

Goal: let workspacer drive either tmux or [gtmux](https://github.com/FyrmForge/gtmux)
as its multiplexer, selected per-workspace. tmux stays the default and its
behavior is unchanged.

## Backend selection

Mirrors the existing `github_backend` pattern exactly:

- `config.MuxBackend` string type + consts `MuxTmux` ("tmux", default) / `MuxGtmux` ("gtmux").
- `Multiplexer MuxBackend yaml:"multiplexer,omitempty"` on `WorkspaceConfig`.
- `workspacer.GetBackend(wc)` factory (mirror of `GetProvider(wc)`), default tmux.

Opt in with one yaml line per workspace: `multiplexer: gtmux`.

## Interface (workspacer/backend.go)

```go
type SessionBackend interface {
	HasSession(name string) bool
	ListSessions() ([]string, error)
	KillSession(name string) error
	CreateSession(spec SessionSpec) error // build DETACHED
	Attach(name string) error             // switch-client if inside a session, else attach
}
```

`SessionSpec{Name, Path, Windows[]}`, `WindowSpec{Name, Layout, Path, Panes[]}`,
`PaneSpec{Command, Size}`. Spec assembly (config → spec, sister windows, vim
file-open args baked into `PaneSpec.Command`) is backend-agnostic and lives in
`tmux.go`'s orchestration. Current-session detection is a free
`CurrentSessionName()` (probes `$GTMUX` then go-tmux) — middleware needs it
before a backend is chosen, so it's not on the interface.

## Files

1. `config/config.go` — `MuxBackend` type/consts + `Multiplexer` field.
2. `workspacer/backend.go` *(new)* — interface, spec types, `GetBackend`, `CurrentSessionName`.
3. `workspacer/tmux_backend.go` *(new)* — `tmuxBackend`: the go-tmux `Configuration.Apply`
   build + `RunCommand`/`resizep`/select, moved out of today's functions. No behavior change.
4. `workspacer/gtmux_backend.go` *(new)* — `gtmuxBackend`: drives the gtmux CLI
   (`new -d`, `run <s> new-window -n -c` / `split-window -c` / `select-layout` /
   `resize-pane -x` / `select-window`, `attach`/`switch-client`, `$GTMUX`).
5. `workspacer/tmux.go` — `StartOrSwitchTo*` / `CloseAll` become thin orchestration:
   assemble a `SessionSpec`, then `be := GetBackend(wc); if be.HasSession(n) { be.Attach } else { be.CreateSession; be.Attach }`.
   Public signatures unchanged → call sites untouched.
6. `cli/middleware.go` — `tmux.GetAttachedSessionName()` → `workspacer.CurrentSessionName()`.

## Ordering

- Steps 1–3, 5, 6 first: **tmux behind the interface, behavior identical.** Build + smoke-test with tmux.
- Then step 4: gtmux backend. Flip a test workspace to `multiplexer: gtmux`, verify build+attach.

## Not doing
`cmd/workspacer_old/` (dead), extra config levels, interface methods beyond the five.
