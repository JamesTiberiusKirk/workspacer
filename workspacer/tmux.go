package workspacer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	gotmux "github.com/jubnzv/go-tmux"
)

func StartOrSwitchToSession(
	wsName string,
	wc config.WorkspaceConfig,
	presets map[string]config.SessionConfig,
	project string,
) {
	sessionName := wsName + "-" + project
	server := new(gotmux.Server)

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

	path := filepath.Join(wc.Path, project)
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}

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
			Name:   strconv.Itoa(i+1) + ": " + w.Name,
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
	err = conf.Apply()
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
		fmt.Println(p.ID)

		if len(panesConfig) <= i {
			continue
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
