package workspacer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/joho/godotenv"
)

func newGitHubClient() *github.Client {
	// add workspace to get the right token
	// temp
	_ = godotenv.Load()

	return github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_AUTH"))
}

func generateBlobURL(result *github.CodeResult, lineBegin, lineEnd int) string {
	return fmt.Sprintf("https://github.com/%s/blob/%s/%s#L%d-L%d",
		*result.Repository.FullName,
		*result.Repository.DefaultBranch,
		*result.Path,
		lineBegin,
		lineEnd,
	)
}

func getRepoDefaultBranch(owner, repoName string) string {
	client := github.NewClient(nil)
	repo, _, err := client.Repositories.Get(context.Background(), owner, repoName)
	if err != nil {
		// handle error
	}
	return repo.GetDefaultBranch()
}

// BUG: so this does not work whats so ever lol
func estimateLineNumbers(fragment *string, match *github.Match) (int, int) {
	lines := strings.Split(*fragment, "\n")
	startLine, endLine := 1, 1
	currentLine := 1
	fragmentStart := match.Indices[0]
	fragmentEnd := match.Indices[1]

	currentPos := 0
	for i, line := range lines {
		if currentPos <= fragmentStart && fragmentStart < currentPos+len(line) {
			startLine = i + 1
		}
		if currentPos <= fragmentEnd && fragmentEnd <= currentPos+len(line) {
			endLine = i + 1
			break
		}
		currentPos += len(line) + 1 // +1 for the newline character
		currentLine++
	}

	return startLine, endLine
}

func SearchGithubInUserOrOrg(userOrOrg, search string) {

	// NOTE: Ok so this does work but i would need to provide

	client := newGitHubClient()
	// client.Search.Repositories(context.Background(), search, nil)

	// i wanna do a global search only inside the user or org
	searchResp, githubResp, err := client.Search.Code(context.Background(), search+" org:"+userOrOrg, &github.SearchOptions{TextMatch: true})
	if err != nil {
		panic(err)
	}
	if githubResp.StatusCode != 200 {
		panic(githubResp.Status)
	}
	// for _, code := range searchResp.CodeResults {
	// 	fmt.Println(code.Repository.Name)
	// 	fmt.Printf("Repo: %s\n", *code.Repository.FullName)
	// 	fmt.Printf("Path: %s\n", *code.Path)
	// 	fmt.Printf("URL: %s\n", *code.HTMLURL)
	// 	fmt.Printf("Text matches: %v\n", code.TextMatches)
	// 	fmt.Printf("", *code.)
	// }

	fmt.Printf("Len: %d, Total: %d\n", len(searchResp.CodeResults), searchResp.GetTotal())

	// b, err := json.MarshalIndent(searchResp, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(b))

	for _, result := range searchResp.CodeResults {

		if result.Repository.DefaultBranch == nil {
			repo, _, err := client.Repositories.Get(context.Background(), *result.Repository.Owner.Login,
				*result.Repository.Name)
			if err != nil {
				panic(err)
			}
			db := repo.GetDefaultBranch()
			result.Repository.DefaultBranch = &db
		}

		fmt.Printf("File: %s\n", *result.Name)

		for _, match := range result.TextMatches {
			lb, le := estimateLineNumbers(match.Fragment, match.Matches[0])
			blobURL := generateBlobURL(result, lb, le)
			fmt.Printf("Blob URL: %s\n", blobURL)

			fmt.Printf("  Match: \n%s\n", *match.Fragment)
		}
		fmt.Println()
	}
}
