package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"

	gin "gopkg.in/gin-gonic/gin.v1"
	githubhook "gopkg.in/rjz/githubhook.v0"
)

func GitHubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func GetCommitPatch(c *github.Client, ctx context.Context, owner, repo, sha string) (string, *github.Response, error) {
	u := fmt.Sprintf("repos/%v/%v/commits/%v", owner, repo, sha)
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return "", nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3.diff")

	patch := new(bytes.Buffer)
	resp, err := c.Do(ctx, req, patch)
	if err != nil {
		return "", resp, err
	}

	return patch.String(), resp, nil
}

func hardcodedSleepAddedIn(patch string) bool {
	addedSleepExpression := "^[+][^#/]*[Ss]leep[ (]+[0-9]+[ )]*"
	match, err := regexp.MatchString(addedSleepExpression, patch)
	if err != nil {
		log.Printf("error matching regex: %s\n", err.Error())
		return false
	}
	return match
}

func main() {
	secretenv := os.Getenv("GITHUB_SECRET")
	if secretenv == "" {
		log.Fatal("$GITHUB_SECRET must be set!")
	}
	secret := []byte(secretenv)

	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("$GITHUB_ACCESS_TOKEN must be set!")
	}

	r := gin.Default()
	gin.SetMode(gin.DebugMode)

	r.POST("/hook", func(c *gin.Context) {
		hook, err := githubhook.Parse(secret, c.Request)
		if err != nil {
			log.Fatalf("error parsing hook: %+v", err.Error())
		}

		eventHeaders := c.Request.Header["X-Github-Event"]
		if len(eventHeaders) == 0 {
			log.Fatalf("could not get X-Github-Event header!")
		}
		eventHeader := eventHeaders[0]
		if eventHeader != "pull_request" {
			log.Printf("event %s is not 'pull_request'\n", eventHeader)
			return
		}

		event := github.PullRequestEvent{}
		err = json.Unmarshal(hook.Payload, &event)
		if err != nil {
			log.Fatalf("error parsing event: %+v\n", err.Error())
		}

		pr := event.PullRequest
		owner := *event.Repo.Owner.Login
		repo := *event.Repo.Name

		switch event.GetAction() {
		case "opened", "edited", "reopened", "synchronize":
		default:
			log.Fatalf("action is %s\n", event.GetAction())
		}

		ctx := context.Background()
		client := GitHubClient(ctx, token)

		name := "insomniac"
		state := "pending"
		status := &github.RepoStatus{Context: &name, State: &state}

		head := pr.Head.GetSHA()
		commit, response, err := client.Git.GetCommit(ctx, owner, repo, head)
		client.Repositories.CreateStatus(ctx, owner, repo, commit.GetSHA(), status)
		log.Printf("setting status to pending")
		if err != nil {
			log.Fatalf("could not get commit %s (%s)\n", head, err.Error())
		}
		if response.StatusCode != http.StatusOK {
			log.Fatalf("response was %d (%+v)\n", response.StatusCode, response)
		}

		for commitsRemaining := pr.GetCommits(); commitsRemaining > 0; commitsRemaining-- {
			head = commit.GetSHA()
			commitPatch, response, err := GetCommitPatch(client, ctx, owner, repo, head)
			if err != nil {
				log.Fatalf("could not get commit %s patch (%s)\n", head, err.Error())
			}
			if response.StatusCode != http.StatusOK {
				log.Fatalf("response was %d (%+v)\n", response.StatusCode, response)
			}

			log.Printf("commit patch is: %s\n", commitPatch)
			if hardcodedSleepAddedIn(commitPatch) {
				log.Printf("discovered hardcoded sleep call")
				*status.State = "failure"
				*status.Description = "no. stop it!"
			} else {
				log.Printf("did not find hardcoded sleep call")
				*status.State = "success"
				*status.Description = "yay!"
			}
			log.Printf("setting status to %s", *status.State)
			client.Repositories.CreateStatus(ctx, owner, repo, commit.GetSHA(), status)
			commit = &commit.Parents[0]
		}
	})

	r.Run()
}
