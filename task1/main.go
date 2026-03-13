package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	timeout = flag.Uint("t", 5, "timeout for HTTP requests (in seconds)")
)

var (
	pattern  = `^(?:https?://github\.com/|git@github\.com:|ssh://git@github\.com/)(?P<owner>[a-zA-Z0-9-._]+)/(?P<repo>[a-zA-Z0-9-._]+)(?:(?:\.git)?(?:/.*)?)$`
	re       = regexp.MustCompile(pattern)
	ownerIdx = re.SubexpIndex("owner")
	repoIdx  = re.SubexpIndex("repo")
)

var client = &http.Client{}

type repository struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	StargazersCount uint      `json:"stargazers_count"`
	ForksCount      uint      `json:"forks_count"`
	CreatedAt       time.Time `json:"created_at"`
}

func (r *repository) String() string {
	description := "<no description>"
	if r.Description != "" {
		description = r.Description
	}
	return fmt.Sprintf("%s: %s\n%d stars\n%d forks\nCreated at: %s",
		r.Name,
		description,
		r.StargazersCount,
		r.ForksCount,
		r.CreatedAt.Format("15:04:05 02-01-2006"),
	)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: pugovkin [flags] [repo ...]\n")
	flag.PrintDefaults()
}

func urlParse(url string) (string, string, error) {
	matches := re.FindStringSubmatch(url)
	if matches == nil {
		return "", "", fmt.Errorf("%s: invalid repository URL", url)
	}

	owner := matches[ownerIdx]
	repo := matches[repoIdx]
	return owner, repo, nil
}

func newRepoRequest(owner, repo string) (*http.Request, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	token, exists := os.LookupEnv("PUGOVKIN_GITHUB_TOKEN")
	if exists {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return req, nil
}

func getRepoData(owner, repo string) (*repository, error) {
	req, err := newRepoRequest(owner, repo)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("repository not found")
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("bad response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data repository
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("invalid response data: %w", err)
	}
	return &data, nil
}

func pugovkin(url string) (string, error) {
	owner, repo, err := urlParse(url)
	if err != nil {
		return "", err
	}

	data, err := getRepoData(owner, repo)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(data), nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	urls := flag.Args()

	if len(urls) == 0 {
		usage()
		os.Exit(1)
	}

	client.Timeout = time.Duration(*timeout) * time.Second
	results := make([]string, len(urls))

	var fetchGroup sync.WaitGroup
	for i, url := range urls {
		fetchGroup.Add(1)
		go func(index int, u string) {
			defer fetchGroup.Done()

			if result, err := pugovkin(u); err != nil {
				results[index] = err.Error() + "\n"
			} else {
				results[index] = result + "\n"
			}
		}(i, url)
	}

	fetchGroup.Wait()

	for _, res := range results {
		fmt.Println(res)
	}
}
