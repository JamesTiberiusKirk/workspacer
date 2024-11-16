package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/log"
)

/*
TODO: REFACTOR THIS
This entire arg and flag implementation is kind of stupid
The whole reason why i wanted it was to be able to have flags specifically for every command.
I.E. -D for debug, -W for workspace is specifically for the base level command
Then say the clone command would have another set of flags
*/

type MainFlags struct {
	Workspace string
	Debug     bool
}

func ParseArgs(args []string) (MainFlags, []string) {
	m := MainFlags{}

	// NOTE: sadly this does not work
	// 0 will always be the binary name and not the alias name sadly
	// TODO: remove this feature, or find a way to make it work
	if !strings.Contains(args[0], "go-build") || strings.Contains(args[0], "workspacer") {
		log.Debug("Getting workspace from args[0]: %s\n", args[0])
		m.Workspace = args[0]
	}

	args = args[1:]
	flaglessArgs := []string{}

	skipNext := false
	mainCommandDone := false
	for i, arg := range args {
		log.Debug("Processing arg: %d: %s: ", i, arg)

		if skipNext {
			log.Debug("Skipping arg processing %d\n", i)
			skipNext = false
			continue
		}

		if !mainCommandDone && strings.HasPrefix(arg, "-") {
			log.Debug("Arg is a flag: %s\n", arg)

			flag := strings.TrimPrefix(arg, "-")
			value := ""
			flagSplit := strings.Split(flag, "=")

			if len(flagSplit) > 1 && flagSplit[1] != "" {
				log.Debug("Flag's value assigned with '=': %v\n", flagSplit)
				flag = flagSplit[0]
				value = flagSplit[1]
			} else {
				log.Debug("Flag's value assigned with next arg\n")
				next := ""
				if i+1 < len(args) {
					next = args[i+1]
				}

				if !strings.HasPrefix(next, "-") {
					value = next
					skipNext = true
				}
			}

			switch flag {
			case "H", "help":
				printHelp()
				os.Exit(0)
			case "W", "workspace":
				log.Debug("args workspace: %s\n", value)
				m.Workspace = value
			case "D", "debug":
				log.Debug("args debug: %s\n", value)
				log.LogLevel = log.LogLevelDebug
				m.Debug = true
			default:
				log.Warn("unknown flag %s %s\n", flag, value)
			}
		} else {
			flaglessArgs = append(flaglessArgs, arg)
			mainCommandDone = true
		}
	}

	return m, flaglessArgs
}

func printHelp() {
	fmt.Printf(`workspacer - a tmux workspace manager with extras

	Run with no parameters to use get project selector

	Avaliable flags for base command:
	-H, -help: print this help message
	-W, -workspace: specify the workspace to urself
	-D, -debug: print debug messages
`)
}
