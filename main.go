package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cli/go-gh"
	"github.com/google/go-github/v69/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	userFlag  string
	tokenFlag string
	since     string
	until     string
	debug     bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gh-nippou",
		Short: "CLI to get GitHub daily report (nippou)",
		Run:   run,
	}

	rootCmd.Flags().StringVarP(&userFlag, "user", "u", "", "GitHub username (optional)")
	rootCmd.Flags().StringVarP(&tokenFlag, "token", "t", "", "GitHub Token (optional)")
	rootCmd.Flags().StringVarP(&since, "since", "s", "", "Start date (YYYYMMDD)")
	rootCmd.Flags().StringVarP(&until, "until", "e", "", "End date (YYYYMMDD)")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Debug mode")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// --- Get token ---
	var ghToken string
	if tokenFlag != "" {
		ghToken = tokenFlag
	} else {
		out, _, err := gh.Exec("auth", "token")
		if err != nil {
			log.Fatalf("failed to get GitHub token: %v", err)
		}
		ghToken = strings.TrimSpace(out.String())
	}

	// --- Get user ---
	var user string
	if userFlag != "" {
		user = userFlag
	} else {
		out, _, err := gh.Exec("api", "user", "--jq", ".login")
		if err != nil {
			log.Fatalf("failed to get GitHub user: %v", err)
		}
		user = strings.TrimSpace(out.String())
	}

	// --- Parse dates ---
	loc := time.Now().Location()
	if since == "" {
		since = time.Now().In(loc).Format("20060102")
	}
	if until == "" {
		until = time.Now().In(loc).Format("20060102")
	}
	if since > until {
		log.Fatalf("since date (%s) is after until date (%s)", since, until)
	}

	sinceTime, err := time.ParseInLocation("20060102", since, loc)
	if err != nil {
		log.Fatalf("failed to parse since date: %v", err)
	}
	untilTime, err := time.ParseInLocation("20060102", until, loc)
	if err != nil {
		log.Fatalf("failed to parse until date: %v", err)
	}
	// Include the entire until day
	untilTime = untilTime.Add(24*time.Hour - time.Nanosecond)

	// --- Create GitHub client ---
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// --- Fetch events ---
	events, err := fetchEvents(ctx, client, user, sinceTime, untilTime)
	if err != nil {
		log.Fatalf("failed to fetch events: %v", err)
	}

	// --- Format output ---
	output := formatOutput(events)
	fmt.Print(output)
}

// Item represents an issue or PR
type Item struct {
	RepoName string
	Title    string
	URL      string
	User     string
	Status   string // "", "merged", "closed"
}

func fetchEvents(ctx context.Context, client *github.Client, user string, sinceTime, untilTime time.Time) ([]Item, error) {
	var allEvents []*github.Event
	opt := &github.ListOptions{Page: 1, PerPage: 100}

	for {
		events, resp, err := client.Activity.ListEventsPerformedByUser(ctx, user, false, opt)
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events...)

		if debug {
			log.Printf("[Debug] Fetched %d events (page %d)", len(events), opt.Page)
		}

		// Check if we should continue
		if resp.NextPage == 0 {
			break
		}
		if len(events) > 0 {
			lastEvent := events[len(events)-1]
			if lastEvent.CreatedAt.Time.Before(sinceTime) {
				break
			}
		}
		opt.Page = resp.NextPage
	}

	// Filter and deduplicate
	seen := make(map[string]bool)
	var items []Item

	for _, event := range allEvents {
		// Check time range
		if event.CreatedAt.Time.Before(sinceTime) || event.CreatedAt.Time.After(untilTime) {
			continue
		}

		item, ok := eventToItem(ctx, client, event)
		if !ok {
			continue
		}

		if debug {
			log.Printf("[Debug] %s: %s %s (%s)", *event.Type, item.RepoName, item.Title, item.Status)
		}

		// Deduplicate by URL
		if seen[item.URL] {
			continue
		}
		seen[item.URL] = true
		items = append(items, item)
	}

	return items, nil
}

func eventToItem(ctx context.Context, client *github.Client, event *github.Event) (Item, bool) {
	payload, err := event.ParsePayload()
	if err != nil {
		return Item{}, false
	}

	repoName := ""
	if event.Repo != nil && event.Repo.Name != nil {
		repoName = *event.Repo.Name
	}

	owner, repo := splitRepoName(repoName)

	switch *event.Type {
	case "IssuesEvent":
		e := payload.(*github.IssuesEvent)
		if e.Issue == nil {
			return Item{}, false
		}
		// Fetch latest issue state from API
		issue, _, err := client.Issues.Get(ctx, owner, repo, e.Issue.GetNumber())
		if err != nil {
			issue = e.Issue // fallback to event data
		}
		status := ""
		if issue.State != nil && *issue.State == "closed" {
			status = "closed"
		}
		return Item{
			RepoName: repoName,
			Title:    getString(issue.Title),
			URL:      getString(issue.HTMLURL),
			User:     getUser(issue.User),
			Status:   status,
		}, true

	case "IssueCommentEvent":
		e := payload.(*github.IssueCommentEvent)
		if e.Issue == nil {
			return Item{}, false
		}
		// Fetch latest issue state from API
		issue, _, err := client.Issues.Get(ctx, owner, repo, e.Issue.GetNumber())
		if err != nil {
			issue = e.Issue // fallback to event data
		}
		status := ""
		if issue.State != nil && *issue.State == "closed" {
			status = "closed"
		}
		return Item{
			RepoName: repoName,
			Title:    getString(issue.Title),
			URL:      getString(issue.HTMLURL),
			User:     getUser(issue.User),
			Status:   status,
		}, true

	case "PullRequestEvent":
		e := payload.(*github.PullRequestEvent)
		if e.PullRequest == nil {
			return Item{}, false
		}
		// Fetch latest PR state from API
		pr, _, err := client.PullRequests.Get(ctx, owner, repo, e.GetNumber())
		if err != nil {
			pr = e.PullRequest // fallback to event data
		}
		status := ""
		if pr.Merged != nil && *pr.Merged {
			status = "merged"
		} else if pr.State != nil && *pr.State == "closed" {
			status = "closed"
		}
		return Item{
			RepoName: repoName,
			Title:    getString(pr.Title),
			URL:      getString(pr.HTMLURL),
			User:     getUser(pr.User),
			Status:   status,
		}, true

	case "PullRequestReviewEvent":
		e := payload.(*github.PullRequestReviewEvent)
		if e.PullRequest == nil {
			return Item{}, false
		}
		// Fetch latest PR state from API
		pr, _, err := client.PullRequests.Get(ctx, owner, repo, e.PullRequest.GetNumber())
		if err != nil {
			pr = e.PullRequest // fallback to event data
		}
		status := ""
		if pr.Merged != nil && *pr.Merged {
			status = "merged"
		} else if pr.State != nil && *pr.State == "closed" {
			status = "closed"
		}
		return Item{
			RepoName: repoName,
			Title:    getString(pr.Title),
			URL:      getString(pr.HTMLURL),
			User:     getUser(pr.User),
			Status:   status,
		}, true

	case "PullRequestReviewCommentEvent":
		e := payload.(*github.PullRequestReviewCommentEvent)
		if e.PullRequest == nil {
			return Item{}, false
		}
		// Fetch latest PR state from API
		pr, _, err := client.PullRequests.Get(ctx, owner, repo, e.PullRequest.GetNumber())
		if err != nil {
			pr = e.PullRequest // fallback to event data
		}
		status := ""
		if pr.Merged != nil && *pr.Merged {
			status = "merged"
		} else if pr.State != nil && *pr.State == "closed" {
			status = "closed"
		}
		return Item{
			RepoName: repoName,
			Title:    getString(pr.Title),
			URL:      getString(pr.HTMLURL),
			User:     getUser(pr.User),
			Status:   status,
		}, true
	}

	return Item{}, false
}

func splitRepoName(repoName string) (owner, repo string) {
	parts := strings.Split(repoName, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getUser(u *github.User) string {
	if u == nil || u.Login == nil {
		return ""
	}
	return *u.Login
}

func formatOutput(items []Item) string {
	if len(items) == 0 {
		return ""
	}

	// Sort by repo name then URL
	sort.Slice(items, func(i, j int) bool {
		if items[i].RepoName != items[j].RepoName {
			return items[i].RepoName < items[j].RepoName
		}
		return items[i].URL < items[j].URL
	})

	var sb strings.Builder
	var prevRepo string

	for _, item := range items {
		if item.RepoName != prevRepo {
			if prevRepo != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("### %s\n\n", item.RepoName))
			prevRepo = item.RepoName
		}

		statusStr := ""
		if item.Status == "merged" {
			statusStr = " **merged!**"
		} else if item.Status == "closed" {
			statusStr = " **closed!**"
		}

		sb.WriteString(fmt.Sprintf("* [%s](%s) by @[%s](https://github.com/%s)%s\n",
			item.Title, item.URL, item.User, item.User, statusStr))
	}

	return sb.String()
}
