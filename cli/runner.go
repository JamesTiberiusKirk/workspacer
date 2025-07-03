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
	// Cursed ik, but it works perfectly
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "-help") ||
		len(os.Args) == 3 && (os.Args[2] == "-h" || os.Args[2] == "-help") ||
		len(os.Args) == 4 && (os.Args[3] == "-h" || os.Args[3] == "-help") {
		printHelp(cm)
		return
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
	fmt.Fprintln(w, "\tworkspacer <global flags> [subCommand] <...>")
	fmt.Fprintln(w, "\tGlobal flags:")
	fmt.Fprintln(w, "\t\t-h,help\tPrint this message.")
	fmt.Fprintln(w, "\t\t-workspace\tDefine workspace.")
	fmt.Fprintln(w)

	// Sort command names case-insensitively
	keys := make([]string, 0, len(cm))
	for k := range cm {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})

	const maxWidth = 60

	fmt.Fprintln(w, "Commands:")
	for _, k := range keys {
		v := cm[k]

		if v.Description != "" {
			descLines := strings.Split(v.Description, "\n")
			wrappedDesc := []string{}
			for _, dl := range descLines {
				wrappedDesc = append(wrappedDesc, wrapText(strings.TrimSpace(dl), maxWidth)...)
			}
			// Print first line with key
			fmt.Fprintf(w, "\t\t%s\t%s\n", k, wrappedDesc[0])
			// Print rest of wrapped description lines indented
			for _, line := range wrappedDesc[1:] {
				fmt.Fprintf(w, "\t\t\t%s\n", line)
			}
		}

		// if v.Help != "" {
		// 	// Blank line before flags/help
		// 	fmt.Fprintln(w, "\t\t\t\t")
		// 	fmt.Fprintf(w, "\t\t\tFlags:")
		//
		// 	helpLines := strings.Split(v.Help, "\n")
		// 	for _, hl := range helpLines {
		// 		wrappedHelp := wrapText(strings.TrimSpace(hl), maxWidth)
		// 		for _, line := range wrappedHelp {
		// 			fmt.Fprintf(w, "\t\t\t%s\n", line)
		// 		}
		// 	}
		// }

		// Blank line after each command block
		// fmt.Fprintln(w, "\t\t\t\t")
	}

	w.Flush()
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
