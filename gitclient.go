package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// GitClient is an interface to return a list of Git tags.
type GitClient interface {
	ListTags() ([]string, error)
}

// GitHubClient is a GitClient that can return a list of tags from github.com for a repo.
type GitHubClient struct {
	client *github.Client
	owner  string
	repo   string
	debug  bool
}

// NewGitHubClient returns a new GitHubClient.
func NewGitHubClient(owner, repo string, debug bool) GitClient {
	var oauth2Client *http.Client

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		oauth2Client = oauth2.NewClient(context.Background(), ts)
	} else {
		if debug {
			fmt.Println("no GITHUB_AUTH_TOKEN env var found so using unauthenticated request")
		}
	}

	return &GitHubClient{
		client: github.NewClient(oauth2Client),
		owner:  owner,
		repo:   repo,
		debug:  debug,
	}
}

// ListTags returns a list of tags from github.com for a repo.
func (g *GitHubClient) ListTags() ([]string, error) {
	if g.debug {
		fmt.Printf("Get tags from github.com/%s/%s\n", g.owner, g.repo)
	}
	ctx := context.Background()
	tags, _, err := g.client.Repositories.ListTags(ctx, g.owner, g.repo, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %v", err)
	}

	var rv []string
	for _, t := range tags {
		rv = append(rv, t.GetName())
	}
	return rv, err
}

// LocalGitClient is a GitClient that can return a list of tags from a local Git repo.
type LocalGitClient struct {
	dir   string
	fetch bool
	debug bool
}

// NewLocalGitClient returns a new LocalGitClient.
func NewLocalGitClient(dir string, fetch, debug bool) GitClient {
	return &LocalGitClient{
		dir:   dir,
		fetch: fetch,
		debug: debug,
	}
}

// ListTags returns a list of tags from a local Git repo.
func (g *LocalGitClient) ListTags() ([]string, error) {
	if g.debug {
		fmt.Printf("Get tags from local repo %s\n", g.dir)
	}

	_, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("error finding git: %v", err)
	}
	if g.fetch {
		cmd := exec.Command("git", "fetch", "--tags", "-v")
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Dir = g.dir
		out, err := cmd.Output()
		if err != nil && g.debug {
			fmt.Printf("ignoring error from `git fetch`: %v\n%v\n", err, out)
		}
	}

	cmd := exec.Command("git", "tag", "--list")
	cmd.Dir = g.dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running `git tag`: %v", err)
	}

	str := strings.TrimSuffix(string(out), "\n")
	tags := strings.Split(str, "\n")
	return tags, nil
}
