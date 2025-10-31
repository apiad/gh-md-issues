# `gh-md-issues`: A Literate Go `gh` Extension

This is the annotated source code for `gh-md-issues`, a `gh` CLI extension to two-way sync GitHub issues with a local folder of markdown files.

This file is a "literate" document. You can read it, or you can "tangle" it into source code using the [illiterate](https://github.com/apiad/illiterate) tool.

## Packages

Our Go project consists of a single file, `main.go`. We'll build this file by exporting several named fragments into one main export block.

This first block is the "root" of our file. It defines the package, imports, and the overall structure of our program.

```go {export=main.go}
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cli/go-gh" // The official 'gh' library
)

```

## Implementation

Now, let's define each of those `<<...>>` fragments one by one.

### Global Constants

First, we'll define our global constants, just like in our shell script. This makes them easy to change later.

```go {export=main.go}
// Global config
const (
    openDir    = "issues"
    closedDir  = "issues/closed"
    stateFile  = ".issues-sync-state"
)

```

### The Main Function

The `main` function is our CLI router. It checks the command-line arguments (`os.Args`) and decides which function to run.

```go {export=main.go}
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

```

### The `pull` Command

This is the stub for our `pull` logic. It uses the `go-gh` library to get the current repository.

```go {export=main.go}
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

```

### The `push` Command

This is the stub for our `push` logic. It shows how you can use `gh.Exec()` to run other commands, like `git status`.

```go {export=main.go}
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

```

### The Usage Function

Finally, a simple helper function to print the help text if the user provides no command or an unknown one.

```go {export=main.go}
// usage prints the help text
func usage() {
    fmt.Println("A tool to two-way sync GitHub issues with local markdown files.")
    fmt.Println("\nUsage: gh md-issues [pull|push]")
    fmt.Println("  pull   Fetches remote issues and updates local files.")
    fmt.Println("  push   Pushes local file changes to remote issues.")
}

```
