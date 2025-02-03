package config

import (
	"errors"
	"slices"
	"strings"
)

type Options struct {
	Workspace string
	Debug     bool
	Help      bool
}

func ParseArgs(args []string) (Options, []error) {
	opts := Options{}
	ignoreList := []string{}
	errorList := []error{}

	for i, a := range args {

		// For skipping next arg if user does not use equal for value assignment
		// E.G. -W ws instead of -W=ws
		if slices.Contains(ignoreList, a) {
			continue
		}

		split := []string{}

		rawSplit := strings.Split(a, "=")
		for _, s := range rawSplit {
			if s == "" {
				continue
			}
			split = append(split, s)
		}

		switch split[0] {
		case "-H", "-help":
			opts.Help = true
		case "-W", "-workspace":
			if len(split) > 1 {
				if strings.Contains(a, "=") && len(split) == 1 {
					errorList = append(errorList, errors.New("Workspace flag needs the name of a workspace"))
					continue
				}
				opts.Workspace = split[1]
			} else {
				if !(len(args) > i+1) || strings.HasPrefix("-", args[i+1]) {
					errorList = append(errorList, errors.New("Workspace flag needs the name of a workspace"))
					continue
				}
				opts.Workspace = args[i+1]
				ignoreList = append(ignoreList, args[i+1])
			}
		case "-D", "-debug":
			opts.Debug = true
		}
	}

	return opts, errorList
}
