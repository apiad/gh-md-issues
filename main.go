package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cli/go-gh" // The official 'gh' library
)

// Global config
const (
    openDir    = "issues"
    closedDir  = "issues/closed"
    stateFile  = ".issues-sync-state"
)

func main() {
    // 1. Check if 'pull' or 'push' was provided
    if len(os.Args) < 2 {
        usage()
        os.Exit(1)
    }

    // 2. Route to the correct function
    command := os.Args[1]
    switch command {
    case "pull":
        if err := pullIssues(); err != nil {
            log.Fatalf("Error during pull: %v", err)
        }
    case "push":
        if err := pushIssues(); err != nil {
            log.Fatalf("Error during push: %v", err)
        }
    default:
        fmt.Printf("Error: unknown command '%s'\n", command)
        usage()
        os.Exit(1)
    }
}

// pullIssues will fetch issues from the GitHub API
func pullIssues() error {
    fmt.Println("Running 'pull'...")

    // Get the repository we are currently in
    repo, err := gh.CurrentRepository()
    if err != nil {
        return fmt.Errorf("could not determine current repository: %w", err)
    }
    fmt.Printf("Operating on repository: %s/%s\n", repo.Owner(), repo.Name())

    // TODO:
    // 1. Read the stateFile to get the last sync time.
    // 2. Build the search query.
    // 3. Use 'gh.RESTClient()' to get an API client.
    // 4. Call client.Get() with the issues endpoint and search query.
    // 5. Parse the JSON response.
    // 6. Loop through issues, format frontmatter, and write files.
    // 7. Write the new timestamp to stateFile.

    fmt.Println("'pull' logic is not yet implemented.")
    return nil
}

// pushIssues will sync local files up to GitHub
func pushIssues() error {
    fmt.Println("Running 'push'...")

    // Get the repository we are currently in
    repo, err := gh.CurrentRepository()
    if err != nil {
        return fmt.Errorf("could not determine current repository: %w", err)
    }
    fmt.Printf("Operating on repository: %s/%s\n", repo.Owner(), repo.Name())

    // TODO:
    // 1. Get list of changed/new/deleted files using 'gh.Exec("git", ...)'
    // 2. Loop through the file list.
    // 3. If deleted, call 'gh.Exec("issue", "close", ...)'
    // 4. If new/modified, parse the frontmatter.
    // 5. If no number, call 'gh.Exec("issue", "create", ...)'
    // 6. If has number, call 'gh.Exec("issue", "edit", ...)'

    fmt.Println("'push' logic is not yet implemented.")
    return nil
}

// usage prints the help text
func usage() {
    fmt.Println("A tool to two-way sync GitHub issues with local markdown files.")
    fmt.Println("\nUsage: gh md-issues [pull|push]")
    fmt.Println("  pull   Fetches remote issues and updates local files.")
    fmt.Println("  push   Pushes local file changes to remote issues.")
}

