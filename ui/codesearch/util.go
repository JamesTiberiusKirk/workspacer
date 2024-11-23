package codesearch

import (
	"context"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/google/go-github/v66/github"
)

type githubCodeSearchFunc func(ctx context.Context, query string, opts *github.SearchOptions) (*github.CodeSearchResult, *github.Response, error)

func highlightCode(query, filterText, code, language string) string {
	hightlightTokens := strings.Split(query, " ")

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}
	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}
	syntaxHightlited := buf.String()

	for _, token := range hightlightTokens {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(token))
		syntaxHightlited = re.ReplaceAllStringFunc(syntaxHightlited, func(match string) string {
			return highlightStyle.Render(match)
		})
	}

	// if filterText != "" {
	// 	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(filterText))
	// 	syntaxHightlited = re.ReplaceAllStringFunc(syntaxHightlited, func(match string) string {
	// 		return filterStyle.Render(match)
	// 	})
	// }

	return syntaxHightlited
}

func wrapText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		wrappedLines = append(wrappedLines, line[0:width-1])
		wl := line[width:]
		for {
			if len(wl) < width {
				wrappedLines = append(wrappedLines, wl)
				break
			}
			wl = wl[:width-1]
			wrappedLines = append(wrappedLines, wl)
		}

	}

	return strings.Join(wrappedLines, "\n")
}

type githubCodeSearchResult struct {
	repo     string
	file     string
	content  string
	language string
}
