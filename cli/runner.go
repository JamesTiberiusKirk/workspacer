package cli

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"
)

func Run(cm ConfigMapType) {
	// parse help out of arg list but NOT after any sub commands
	for _, a := range os.Args[1:] {
		if !strings.HasPrefix(a, "-") {
			break
		}
		if a == "-h" || a == "-help" {
			printHelp(cm)
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	args := os.Args[1:]
	customCtx := ConfigMapCtx{
		Context: ctx,
		Args:    args,
	}

	subCommand := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++ // skip the value
			}
			continue
		}

		subCommand = arg
		break
	}

	for k, v := range cm {
		split := strings.Split(k, ",")
		if !slices.Contains(split, subCommand) {
			continue
		}

		v.Runner(customCtx)
		return
	}

	if subCommand == "" {
		subCommand = CommandTypeNoCommand
	}

	cm[subCommand].Runner(customCtx)
}

func printHelp(cm ConfigMapType) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "Usage:\t")

	projectCommand, ok := cm[CommandTypeDefault]
	if ok {
		printCommand(w, "workspacer -W workspace <project>", projectCommand.Description)
	}

	noCommand, ok := cm[CommandTypeNoCommand]
	if ok {
		printCommand(w, "workspacer -W workspace", noCommand.Description)
	}

	printCommand(w, "workspacer <global flags> [subCommand] <...>", "")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "\tGlobal flags:")
	fmt.Fprintln(w, "\t\t-h,help\tPrint this message.")
	fmt.Fprintln(w, "\t\t-workspace\tDefine workspace.")
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	// Sort command names case-insensitively
	keys := make([]string, 0, len(cm))
	for k := range cm {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})

	fmt.Fprintln(w, "Sub commands:")
	for _, k := range keys {
		v := cm[k]

		if v.Description == CommandTypeDefault {
			continue
		}

		if v.Description == CommandTypeNoCommand {
			continue
		}

		if v.Description != "" {
			printCommand(w, k, v.Description)
		}
	}

	w.Flush()
}

const maxWidth = 60

func printCommand(w *tabwriter.Writer, name, desc string) {
	descLines := strings.Split(desc, "\n")
	wrappedDesc := []string{}
	for _, dl := range descLines {
		wrappedDesc = append(wrappedDesc, wrapText(strings.TrimSpace(dl), maxWidth)...)
	}
	// Print first line with key
	fmt.Fprintf(w, "\t\t%s\t%s\n", name, wrappedDesc[0])
	// Print rest of wrapped description lines indented
	for _, line := range wrappedDesc[1:] {
		fmt.Fprintf(w, "\t\t\t%s\n", line)
	}
}

func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	indent := "  "
	indentLen := len(indent)

	lines := []string{}
	line := words[0]
	isFirstLine := true

	for _, w := range words[1:] {
		limit := maxWidth
		if !isFirstLine {
			limit -= indentLen
		}

		if len(line)+1+len(w) > limit {
			if isFirstLine {
				lines = append(lines, line)
				line = indent + w
				isFirstLine = false
			} else {
				lines = append(lines, line)
				line = indent + w
			}
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return lines
}
