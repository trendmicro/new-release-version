package adapters

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

type GitClient interface {
	ListTags() ([]string, error)
}

type GitHubClient struct {
	client *github.Client
	owner  string
	repo   string
	debug  bool
}

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

type LocalGitClient struct {
	dir   string
	debug bool
}

func NewLocalGitClient(dir string, debug bool) GitClient {
	return &LocalGitClient{
		dir:   dir,
		debug: debug,
	}
}

func (g *LocalGitClient) ListTags() ([]string, error) {
	if g.debug {
		fmt.Printf("Get tags from local repo %s\n", g.dir)
	}

	_, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("error finding git: %v", err)
	}
	cmd := exec.Command("git", "fetch", "--tags", "-v")
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Dir = g.dir
	out, err := cmd.Output()
	if err != nil && g.debug {
		fmt.Printf("ignoring error from `git fetch`: %v\n%v\n", err, out)
	}

	cmd = exec.Command("git", "tag", "--list")
	cmd.Dir = g.dir
	out, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running `git tag`: %v\n", err)
	}

	str := strings.TrimSuffix(string(out), "\n")
	tags := strings.Split(str, "\n")
	if g.debug {
		fmt.Printf("found tags: %v\n", tags)
	}

	return tags, nil
}
