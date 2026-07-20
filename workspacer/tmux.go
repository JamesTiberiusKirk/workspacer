package workspacer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/util"
)

// sanitizeTmuxName replaces characters multiplexers disallow in session names.
func sanitizeTmuxName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

// applyVimArgs appends the project's file/extra-command options to a vim-family
// pane command (from the `project:file:extra` target syntax).
func applyVimArgs(cmd, fileOption, extraVimCommands string) string {
	switch cmd {
	case "vi", "vim", "nvim":
		if fileOption != "" {
			cmd += " ./" + fileOption
		}
		if extraVimCommands != "" {
			cmd += " " + extraVimCommands
		}
	}
	return cmd
}

// StartOrSwitchToTmpSession creates (or attaches/switches to) a session named
// after the given path and rooted in it. No workspace or preset — always tmux
// (there's no workspace config to select a backend from).
func StartOrSwitchToTmpSession(path string) {
	name := sanitizeTmuxName(filepath.Base(path))
	be := GetBackend()

	if !be.HasSession(name) {
		spec := SessionSpec{Name: name, Path: path, Windows: []WindowSpec{{Panes: []PaneSpec{{}}}}}
		if err := be.CreateSession(spec); err != nil {
			fmt.Printf("Error creating session: %s\n", err.Error())
			return
		}
	}
	if err := be.Attach(name); err != nil {
		fmt.Printf("Error attaching to session: %s\n", err.Error())
	}
}

func CloseAllSessionsInWorkspace(wc config.WorkspaceConfig) {
	if wc.Prefix == "" {
		fmt.Println("prefix is empty")
		return
	}
	be := GetBackend()
	names, err := be.ListSessions()
	if err != nil {
		fmt.Println("error ", err.Error())
		return
	}
	for _, n := range names {
		if strings.HasPrefix(n, wc.Prefix) {
			if err := be.KillSession(n); err != nil {
				fmt.Println("error ", err.Error())
			}
		}
	}
}

// StartOrSwitchToTmuxPreset builds (or attaches to) a session from a standalone
// preset rooted at basePath. No workspace config → always tmux.
func StartOrSwitchToTmuxPreset(name string, basePath string, preset config.SessionConfig) {
	name = sanitizeTmuxName(name)
	be := GetBackend()

	if be.HasSession(name) {
		if err := be.Attach(name); err != nil {
			fmt.Println("error ", err.Error())
		}
		return
	}

	basePath, err := util.ExpandTilde(basePath)
	if err != nil {
		fmt.Println("Error expanding user home folder")
		return
	}

	spec := SessionSpec{Name: name, Path: basePath}
	for _, w := range preset.Windows {
		wp := w.Path
		if wp == "" {
			wp = basePath
		} else if expanded, err := util.ExpandTilde(wp); err == nil {
			wp = expanded
		} else {
			fmt.Printf("Error expanding path %s: %s\n", w.Path, err)
			return
		}
		ws := WindowSpec{Name: w.Name, Layout: w.Layout, Path: wp}
		for _, p := range w.Panes {
			ws.Panes = append(ws.Panes, PaneSpec{Command: p.Command, Size: p.Size})
		}
		spec.Windows = append(spec.Windows, ws)
	}

	if err := be.CreateSession(spec); err != nil {
		fmt.Println(err)
		return
	}
	if err := be.Attach(name); err != nil {
		fmt.Println("error ", err.Error())
	}
}

func StartOrSwitchToSession(
	wc config.WorkspaceConfig,
	presets map[string]config.SessionConfig,
	project string,
) {
	fileOption := ""
	extraVimCommands := ""
	if strings.Contains(project, ":") {
		split := strings.Split(project, ":")
		project = split[0]
		fileOption = split[1]
		if len(split) > 2 {
			extraVimCommands = split[2]
		}
	}

	rootMode := project == ""
	if rootMode {
		project = "root"
		if _, err := os.Stat(util.GetWorkspacePath(wc)); os.IsNotExist(err) {
			fmt.Printf("\n\nWorkspace root does not exist\n\n")
			return
		}
	} else if !util.DoesProjectExist(wc, project) {
		fmt.Printf("\n\nProject %s does not exist\n\n", project)
		return
	}

	sessionName := sanitizeTmuxName(project)
	if wc.Prefix != "" {
		sessionName = sanitizeTmuxName(wc.Prefix) + "-" + sessionName
	}

	be := GetBackend()
	if be.HasSession(sessionName) {
		if err := be.Attach(sessionName); err != nil {
			fmt.Println("error ", err.Error())
		}
		return
	}

	path := util.GetWorkspacePath(wc)
	if !rootMode {
		path = filepath.Join(path, project)
	}

	var sessionConfig config.SessionConfig
	if wc.SessionPreset != "" {
		sessionConfig = presets[wc.SessionPreset]
	} else if wc.Session != nil {
		sessionConfig = *wc.Session
	}

	spec := SessionSpec{Name: sessionName, Path: path}

	// Main windows (first window's name is overridden with the project name).
	for i, w := range sessionConfig.Windows {
		wname := w.Name
		if i == 0 {
			wname = project
		}
		ws := WindowSpec{Name: wname, Layout: w.Layout}
		for _, p := range w.Panes {
			ws.Panes = append(ws.Panes, PaneSpec{
				Command: applyVimArgs(p.Command, fileOption, extraVimCommands),
				Size:    p.Size,
			})
		}
		spec.Windows = append(spec.Windows, ws)
	}

	// Sister repos add windows rooted in their own paths.
	for _, sr := range util.GetSisterReposForProject(wc, project) {
		if !util.DoesProjectExist(wc, sr.Name) {
			continue
		}
		sisterPath := filepath.Join(util.GetWorkspacePath(wc), sr.Name)

		var sisterCfg config.SessionConfig
		if sr.SessionPreset != "" {
			if p, ok := presets[sr.SessionPreset]; ok {
				sisterCfg = p
			}
		}

		if len(sisterCfg.Windows) > 0 {
			for i, w := range sisterCfg.Windows {
				wname := w.Name
				if i == 0 {
					wname = sr.Label
				}
				ws := WindowSpec{Name: wname, Layout: w.Layout, Path: sisterPath}
				for _, p := range w.Panes {
					ws.Panes = append(ws.Panes, PaneSpec{Command: p.Command, Size: p.Size})
				}
				spec.Windows = append(spec.Windows, ws)
			}
		} else {
			spec.Windows = append(spec.Windows, WindowSpec{
				Name:   sr.Label,
				Layout: "even-horizontal",
				Path:   sisterPath,
				Panes:  []PaneSpec{{}},
			})
		}
	}

	if err := be.CreateSession(spec); err != nil {
		fmt.Println(err)
		return
	}
	if err := be.Attach(sessionName); err != nil {
		fmt.Println("error ", err.Error())
	}
}
