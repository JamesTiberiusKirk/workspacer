package config

import (
	"fmt"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/log"
)

type MainFlags struct {
	Workspace string
	Debug     bool
}

func ParseArgs(args []string) (MainFlags, []string) {
	m := MainFlags{
		Workspace: args[0],
	}

	if strings.Contains(args[0], "go-build") {
		log.Debug("running with go run looking for flag, %s", args[0])
	}

	args = args[1:]
	flaglessArgs := []string{}

	skipNext := false
	for i, arg := range args {
		log.Debug("%d: %s: ", i, arg)

		if skipNext {
			log.Debug("skipping %d\n", i)
			continue
		}

		if strings.HasPrefix(arg, "-") {

			flag := strings.TrimPrefix(arg, "-")
			value := ""
			flagSplit := strings.Split(flag, "=")

			if len(flagSplit) > 1 && flagSplit[1] != "" {
				flag = flagSplit[0]
				value = flagSplit[1]
			} else {
				if i+1 <= len(args) && !strings.HasPrefix(args[i], "-") {
					value = args[i+1]
					skipNext = true
				}
			}

			log.Debug("flag: %s, value: %s, args: %v", flag, value, args)

			switch flag {
			case "h", "help":
				fmt.Printf("go help urself\n")
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
		}
	}

	return m, flaglessArgs
}
