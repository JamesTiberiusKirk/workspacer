package workspacer

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/util"
	gotmux "github.com/jubnzv/go-tmux"
)

func CloseAllSessionsInWorkspace(wc config.WorkspaceConfig) {
	server := new(gotmux.Server)
	sessions, err := server.ListSessions()
	if err != nil {
		// handle error
		fmt.Println("error ", err.Error())
	}

	for _, s := range sessions {
		if wc.Prefix == "" {
			fmt.Println("prefix is empty")
			return
		}

		if strings.HasPrefix(s.Name, wc.Prefix) {
			err := server.KillSession(s.Name)
			if err != nil {
				// handle error
				fmt.Println("error ", err.Error())
			}
		}
	}
}

func StartOrSwitchToTmuxPreset(name string, basePath string, preset config.SessionConfig) {
	server := new(gotmux.Server)
	sessions, _ := server.ListSessions()
	if len(sessions) > 0 {
		// Check that the "example" session already exists.
		exists, err := server.HasSession(name)
		if err != nil {
			fmt.Println(fmt.Errorf("Can't check '%s' session: %s", name, err))
			return
		}

		if exists {
			sessions, err := server.ListSessions()
			if err != nil {
				// handle error
				fmt.Println("error ", err.Error())
			}

			for _, s := range sessions {
				if s.Name != name {
					continue
				}
				s.AttachSession()
				break
			}
			return
		}
	}

	windows := []gotmux.Window{}
	for i, w := range preset.Windows {
		panes := []gotmux.Pane{}
		for range w.Panes {
			pane := gotmux.Pane{}
			panes = append(panes, pane)
		}

		window := gotmux.Window{
			Id:     i + 1,
			Name:   w.Name,
			Layout: w.Layout,
			Panes:  panes,
		}

		windows = append(windows, window)
	}

	session := gotmux.NewSession(0, name, basePath, windows)

	server.AddSession(*session)
	conf := gotmux.Configuration{
		Server:        server,
		Sessions:      []*gotmux.Session{session},
		ActiveSession: nil,
	}

	// Setup this configuration.
	err := conf.Apply()
	if err != nil {
		msg := fmt.Errorf("Can't apply prepared configuration: %s", err)
		fmt.Println(msg)
		return
	}

	panes, err := session.ListPanes()
	if err != nil {
		fmt.Println("error ", err.Error())
	}

	panesConfig := preset.ListPanes()
	for i, p := range panes {
		if len(panesConfig) <= i {
			continue
		}

		if panesConfig[i].Command == "vi" ||
			panesConfig[i].Command == "vim" ||
			panesConfig[i].Command == "nvim" {
		}

		if panesConfig[i].Size > 0 && panesConfig[i].Size < 100 {
			paneSize(panesConfig[i].Size)
		}

		p.RunCommand(panesConfig[i].Command)
	}

	{
		// NOTE: Select first window
		windows, err := session.ListWindows()
		if err != nil {
			fmt.Println("error ", err.Error())
		}
		windows[0].Select()
		panes[0].Select()
	}

	// Attach to the created session
	err = session.AttachSession()
	if err != nil {
		msg := fmt.Errorf("Can't attached to created session: %s", err)
		fmt.Println(msg)
		return
	}
}

func StartOrSwitchToSession(
	wsName string,
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

	if !util.DoesProjectExist(wc, project) {
		fmt.Printf("\n\nProject %s does not exist\n\n", project)
		return
	}

	sessionName := project

	if wc.Prefix != "" {
		sessionName = wsName + "-" + sessionName
	}

	server := new(gotmux.Server)

	sessions, _ := server.ListSessions()
	if len(sessions) > 0 {
		// Check that the "example" session already exists.
		exists, err := server.HasSession(sessionName)
		if err != nil {
			fmt.Println(fmt.Errorf("Can't check '%s' session: %s", sessionName, err))
			return
		}

		if exists {
			sessions, err := server.ListSessions()
			if err != nil {
				// handle error
				fmt.Println("error ", err.Error())
			}

			for _, s := range sessions {
				if s.Name != sessionName {
					continue
				}
				s.AttachSession()
				break
			}
			return
		}
	}

	path := filepath.Join(util.GetWorkspacePath(wc), project)

	// TODO: check if the path is valid
	// This is where the project will be cloned if config has been setup

	sessionConfig := wc.Session
	if wc.SessionPreset != "" {
		sessionConfig = presets[wc.SessionPreset]
	}

	windows := []gotmux.Window{}

	for i, w := range sessionConfig.Windows {
		panes := []gotmux.Pane{}
		for range w.Panes {
			pane := gotmux.Pane{}
			panes = append(panes, pane)
		}

		window := gotmux.Window{
			Id:     i + 1,
			Name:   w.Name,
			Layout: w.Layout,
			Panes:  panes,
		}

		windows = append(windows, window)
	}

	session := gotmux.NewSession(0, sessionName, path, windows)

	server.AddSession(*session)
	conf := gotmux.Configuration{
		Server:        server,
		Sessions:      []*gotmux.Session{session},
		ActiveSession: nil,
	}

	// Setup this configuration.
	err := conf.Apply()
	if err != nil {
		msg := fmt.Errorf("Can't apply prepared configuration: %s", err)
		fmt.Println(msg)
		return
	}

	panes, err := session.ListPanes()
	if err != nil {
		fmt.Println("error ", err.Error())
	}

	panesConfig := sessionConfig.ListPanes()
	for i, p := range panes {
		// fmt.Println(p.ID)

		if len(panesConfig) <= i {
			continue
		}

		if panesConfig[i].Command == "vi" ||
			panesConfig[i].Command == "vim" ||
			panesConfig[i].Command == "nvim" {
			if fileOption != "" {
				panesConfig[i].Command += " ./" + fileOption
			}
			if extraVimCommands != "" {
				panesConfig[i].Command += " " + extraVimCommands
			}
		}

		if panesConfig[i].Size > 0 && panesConfig[i].Size < 100 {
			paneSize(panesConfig[i].Size)
		}

		p.RunCommand(panesConfig[i].Command)
	}

	{
		// NOTE: Select first window
		windows, err := session.ListWindows()
		if err != nil {
			fmt.Println("error ", err.Error())
		}
		windows[0].Select()
		panes[0].Select()
	}

	// Attach to the created session
	err = session.AttachSession()
	if err != nil {
		msg := fmt.Errorf("Can't attached to created session: %s", err)
		fmt.Println(msg)
		return
	}
}

func paneSize(size int) error {
	args := []string{
		"resizep",
		// "-t", fmt.Sprintf("%s:%s", w.SessionName, w.Name),
		"-t{right}",
		"-x " + fmt.Sprint(size) + "%"}
	s, err_cmd, err_exec := RunCmd(args)
	if err_exec != nil {
		//HANDLE
		fmt.Print(err_exec.Error())
	}
	if err_cmd != "" {
		//HANDLE
		fmt.Print(err_cmd)
	}
	fmt.Print(s)

	return nil
}

func RunCmd(args []string) (string, string, error) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return "", "", err
	}
	fmt.Print(tmux, fmt.Sprint(args))
	cmd := exec.Command(tmux, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())

	return outStr, errStr, err
}
