package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cli/go-gh"
	"github.com/masutaka/github-nippou/v4/lib"
	"github.com/spf13/cobra"
)

var (
	userFlag     string
	tokenFlag    string
	settingsGist string
	since        string
	until        string
	debug        bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gh-nippou",
		Short: "CLI to get GitHub daily report (nippou)",
		Run:   run,
	}

	// Flag definitions
	rootCmd.Flags().StringVarP(&userFlag, "user", "u", "", "GitHub username (optional)")
	rootCmd.Flags().StringVarP(&tokenFlag, "token", "t", "", "GitHub Token (optional)")
	rootCmd.Flags().StringVarP(&settingsGist, "settings-gist-id", "g", "", "Gist ID for settings")
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

	// get science date and until date if not set
	// should be YYYYMMDD
	loc := time.Now().Location()
	if since == "" {
		since = time.Now().In(loc).Format("20060102")
	}
	if until == "" {
		until = time.Now().In(loc).Format("20060102")
	}
	// check if since date is before until date
	if since > until {
		log.Fatalf("since date (%s) is after until date (%s)", since, until)
	}

	// --- Get nippou ---
	list := lib.NewList(since, until, user, ghToken, settingsGist, debug)
	lines, err := list.Collect()
	if err != nil {
		log.Fatalf("failed to collect nippou: %v", err)
	}

	fmt.Println(lines)
}
