package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var mergeRequest = regexp.MustCompile(`Merge pull request #([\d]+)`)
var owner = "apple"
var repo = "swift"

func main() {
	// logrus.SetLevel(logrus.DebugLevel)
	base, head := getArgs()
	ctx := context.Background()
	client := newGithubClient(ctx)
	commitsComparison := getCommitsComparison(client, ctx, base, head)
	sanitizedCommits := sanitizedCommits(client, ctx, commitsComparison)

	if err := writeListToPath(sanitizedCommits, "./CHANGELOG.md"); err != nil {
		logrus.WithError(err).Fatal("Could not write changelog")
	}
	logrus.Info("🚀🚀🚀 Finised with success 🚀🚀🚀")
	logrus.Info("Check the output at ./CHANGELOG.md")
}

func sanitizedCommits(client *github.Client, ctx context.Context, repositoryCommits []github.RepositoryCommit) []string {
	logrus.Info("Searching for PR's that begins with \"Merge pull request # \" ")
	sanitizedCommits := []string{}
	for i, repoCommit := range repositoryCommits {
		line := "LINE " + strconv.Itoa(i) + ": "
		pullNumber, err := getPullRequestNumber(repoCommit)
		if err != nil {
			logrus.WithError(err).Debug("Saving commit message")
			sanitizedCommits = append(sanitizedCommits, line+repoCommit.Commit.GetMessage())
		} else {
			logrus.Info("Automatic merge found. Fetching PR #", pullNumber)
			pullRequest := getPR(client, ctx, pullNumber)
			sanitizedCommits = append(sanitizedCommits, line+pullRequest.GetTitle())
		}
	}
	return sanitizedCommits
}

func getCommitsComparison(client *github.Client, ctx context.Context, base string, head string) []github.RepositoryCommit {
	commitsComparison, _, err := client.Repositories.CompareCommits(ctx, owner, repo, base, head)
	if err != nil {
		message := fmt.Sprintf("could not fetch the list of commits for HEAD: %s and BASE: %s", head, base)
		logrus.WithError(err).Fatal(message)
	}
	logComparisonInfo(commitsComparison)

	if commitsComparison.GetTotalCommits() == 0 {
		message := fmt.Sprintf("No commits found between HEAD: %s and BASE: %s", head, base)
		logrus.Fatal(message)
	}

	return commitsComparison.Commits
}

func getPR(client *github.Client, ctx context.Context, pullNumber int) *github.PullRequest {
	pullRequest, _, err := client.PullRequests.Get(ctx, owner, repo, pullNumber)
	if err != nil {
		logrus.WithError(err).Fatal("could not fetch PR: #", pullNumber)
	}
	return pullRequest
}

func getPullRequestNumber(repoCommit github.RepositoryCommit) (int, error) {
	message := repoCommit.Commit.GetMessage()
	matches := mergeRequest.FindStringSubmatch(message)
	if len(matches) == 0 {
		return 0, errors.New("This was probably merged by a human. SHA: " + repoCommit.GetSHA())
	}

	pullNumber, err := strconv.Atoi(matches[len(matches)-1])
	if err != nil {
		logrus.Error()
		return 0, errors.New("Weird. Could not convert matched result to number: " + strings.Join(matches, " "))
	}

	return pullNumber, nil
}

func newGithubClient(ctx context.Context) *github.Client {
	token, exists := os.LookupEnv("GITHUB_API_TOKEN")
	if !exists {
		logrus.Warn("GITHUB_API_TOKEN Not provided. Creating unauthenticated client.")
		logrus.Warn("You can export GITHUB_API_TOKEN env to have authenticated client with more limits")
		return github.NewClient(nil)
	}
	logrus.Info("GITHUB_API_TOKEN Provided. Creating authenticated client.")
	logrus.Debug("GITHUB_API_TOKEN: " + token)

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, tokenSource)

	return github.NewClient(tc)
}

func getArgs() (string, string) {
	if len(os.Args) != 3 {
		logrus.Fatal("Please specify BASE and HEAD tags")
	}
	base := os.Args[1]
	head := os.Args[2]
	message := fmt.Sprintf("Running sanitizer between HEAD: %s and BASE: %s", head, base)
	logrus.Info(message)
	return base, head
}

func logComparisonInfo(c *github.CommitsComparison) {
	logrus.Info("##################################")
	logrus.Info("####### Status: ", c.GetStatus())
	logrus.Info("####### AheadBy: ", c.GetAheadBy())
	logrus.Info("####### BehindBy: ", c.GetBehindBy())
	logrus.Info("####### TotalCommits: ", c.GetTotalCommits())
	logrus.Info("##################################")
}

func writeListToPath(list []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range list {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
