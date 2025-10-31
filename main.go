package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp" // --- ADDED ---
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh"
)

// --- Global Configuration ---

const (
	openDir   = "issues"
	closedDir = "issues/closed"
	stateFile = ".issues-sync-state"
)

// --- Structs for API Parsing ---

// Issue represents the data we get from the GitHub API
type Issue struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	Body   string  `json:"body"`
	State  string  `json:"state"`
	Labels []Label `json:"labels"`
}

// Label is a sub-struct for the issue labels
type Label struct {
	Name string `json:"name"`
}

// --- Main CLI Router ---

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

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

// --- PULL Command Logic ---

// pullIssues fetches issues from the GitHub API
func pullIssues() error {
	fmt.Println("Running 'pull'...")

	repo, err := gh.CurrentRepository()
	if err != nil {
		return fmt.Errorf("could not determine current repository: %w", err)
	}
	repoFullName := fmt.Sprintf("%s/%s", repo.Owner(), repo.Name())
	fmt.Printf("Operating on repository: %s\n", repoFullName)

	lastSync, err := getLastSyncTime()
	if err != nil {
		return fmt.Errorf("could not read state file: %w", err)
	}

	newSyncTime := time.Now().UTC()

	issues, err := fetchUpdatedIssues(repoFullName, lastSync)
	if err != nil {
		return fmt.Errorf("could not fetch issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No new issues found. Sync complete.")
		return writeNewSyncTime(newSyncTime)
	}

	fmt.Printf("Found %d issues to update...\n", len(issues))
	if err := processPulledIssues(issues); err != nil {
		return fmt.Errorf("could not write issue files: %w", err)
	}

	return writeNewSyncTime(newSyncTime)
}

func getLastSyncTime() (time.Time, error) {
	data, err := ioutil.ReadFile(stateFile)
	if os.IsNotExist(err) {
		fmt.Println("No state file found. Performing full sync...")
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	lastSync, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		fmt.Println("Warning: could not parse state file. Performing full sync.")
		return time.Time{}, nil
	}

	fmt.Printf("Pulling issues updated since %s\n", lastSync.Format(time.RFC3339))
	return lastSync, nil
}

func fetchUpdatedIssues(repoFullName string, lastSync time.Time) ([]Issue, error) {
	client, err := gh.RESTClient(nil)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("repos/%s/issues?state=all&per_page=100", repoFullName)
	if !lastSync.IsZero() {
		path = fmt.Sprintf("%s&since=%s", path, lastSync.Format(time.RFC3339))
	}

	var issues []Issue
	err = client.Get(path, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

func processPulledIssues(issues []Issue) error {
	if err := os.MkdirAll(openDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(closedDir, 0755); err != nil {
		return err
	}

	for _, issue := range issues {
		var b strings.Builder
		b.WriteString("---\n")
		fmt.Fprintf(&b, "number: %d\n", issue.Number)
		fmt.Fprintf(&b, "title: \"%s\"\n", strings.ReplaceAll(issue.Title, "\"", "\\\""))
		fmt.Fprintf(&b, "state: %s\n", issue.State)

		b.WriteString("labels:\n")
		for _, label := range issue.Labels {
			fmt.Fprintf(&b, "- %s\n", label.Name)
		}
		b.WriteString("---\n\n")
		b.WriteString(strings.ReplaceAll(issue.Body, "\r", ""))

		// --- MODIFIED ---
		slug := generateSlug(issue.Title)
		newFileName := fmt.Sprintf("%d-%s.md", issue.Number, slug)

		var targetDir, oldDir string
		if issue.State == "open" {
			targetDir = openDir
			oldDir = closedDir
		} else {
			targetDir = closedDir
			oldDir = openDir
		}

		// --- REVISED FILE REMOVAL LOGIC ---

		// 1. Clean up the OTHER directory (for state changes)
		os.Remove(filepath.Join(oldDir, fmt.Sprintf("%d.md", issue.Number))) // Remove old "NUMBER.md"
		matchesOld, _ := filepath.Glob(filepath.Join(oldDir, fmt.Sprintf("%d-*.md", issue.Number)))
		for _, match := range matchesOld {
			os.Remove(match)
		}

		// 2. Clean up the TARGET directory (for title/slug changes)
		// This finds any file with the same number but a different slug
		matchesTarget, _ := filepath.Glob(filepath.Join(targetDir, fmt.Sprintf("%d-*.md", issue.Number)))
		for _, match := range matchesTarget {
			if filepath.Base(match) != newFileName {
				fmt.Printf("Removing old file: %s\n", match)
				os.Remove(match)
			}
		}
		// Also remove old "NUMBER.md" format from target dir
		os.Remove(filepath.Join(targetDir, fmt.Sprintf("%d.md", issue.Number)))
		// --- END REVISED LOGIC ---

		filePath := filepath.Join(targetDir, newFileName)
		// --- END MODIFIED ---

		fmt.Printf("Writing file: %s\n", filePath)
		if err := ioutil.WriteFile(filePath, []byte(b.String()), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}
	return nil
}

func writeNewSyncTime(syncTime time.Time) error {
	timestamp := syncTime.Format(time.RFC3339)
	fmt.Printf("Sync complete. Updating state file to %s\n", timestamp)
	return ioutil.WriteFile(stateFile, []byte(timestamp), 0644)
}

// --- PUSH Command Logic ---

// pushIssues will sync local files up to GitHub
func pushIssues() error {
	fmt.Println("Running 'push'...")

	repo, err := gh.CurrentRepository()
	if err != nil {
		return fmt.Errorf("could not determine current repository: %w", err)
	}
	repoFullName := fmt.Sprintf("%s/%s", repo.Owner(), repo.Name())

	if err := os.MkdirAll(openDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(closedDir, 0755); err != nil {
		return err
	}

	// 1. Get list of changed/new/deleted files using 'git status'
	cmd := exec.Command("git", "status", "--porcelain", openDir, closedDir)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run 'git status': %w", err)
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println("No modified files found. Nothing to push.")
		return nil
	}

	// 2. Loop through the file list
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		status := parts[0]
		file := parts[1]

		if !strings.HasSuffix(file, ".md") {
			continue
		}

		fmt.Printf("---\nProcessing %s (Status: %s)\n", file, status)

		// 3. Handle Deletions
		if status == "D" {
			if err := handleDeletedFile(file, repoFullName); err != nil {
				fmt.Printf("Error handling deleted file %s: %v\n", file, err)
			}
		} else {
			// 4. Handle Create/Update
			if err := handleModifiedFile(file, repoFullName); err != nil {
				fmt.Printf("Error handling modified file %s: %v\n", file, err)
			}
		}
	}
	fmt.Println("---")
	fmt.Println("Push complete.")
	return nil
}

// handleDeletedFile closes the corresponding issue on GitHub
func handleDeletedFile(file string, repoFullName string) error {
	fmt.Printf("File %s was deleted.\n", file)
	base := filepath.Base(file)

	// --- MODIFIED ---
	fileNameNoExt := strings.TrimSuffix(base, ".md")
	numberStr := fileNameNoExt
	if parts := strings.SplitN(fileNameNoExt, "-", 2); len(parts) > 0 {
		numberStr = parts[0] // Take only the part before the first hyphen
	}
	// --- END MODIFIED ---

	if _, err := strconv.Atoi(numberStr); err != nil {
		fmt.Printf("Skipping deleted file %s (name is not an issue number).\n", file)
		return nil
	}

	// Check if issue is already closed
	var currentState struct {
		State string `json:"state"`
	}
	client, err := gh.RESTClient(nil)
	if err != nil {
		return err
	}
	err = client.Get(fmt.Sprintf("repos/%s/issues/%s", repoFullName, numberStr), &currentState)
	if err != nil {
		return fmt.Errorf("could not get state for issue #%s: %w", numberStr, err)
	}

	if currentState.State == "open" {
		fmt.Printf("Closing Issue #%s on GitHub...\n", numberStr)
		_, stdErr, err := gh.Exec("issue", "close", numberStr, "-R", repoFullName)
		if err != nil {
			return fmt.Errorf("failed to close issue: %s", stdErr.String())
		}
	} else {
		fmt.Printf("Issue #%s is already closed.\n", numberStr)
	}
	return nil
}

// handleModifiedFile creates or updates an issue on GitHub
// handleModifiedFile creates or updates an issue on GitHub
func handleModifiedFile(file string, repoFullName string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	meta := parseFrontmatter(data)
	body := parseBody(data)

	number, ok := meta["number"]

	if !ok || number == "" {
		// 5. CREATE Issue
		title, ok := meta["title"]
		if !ok || title == "" {
			return fmt.Errorf("new file %s is missing a 'title' in its frontmatter", file)
		}

		fmt.Printf("Creating new issue with title: %s\n", title)

		stdOut, stdErr, err := gh.Exec("issue", "create",
			"-R", repoFullName,
			"-t", title,
			"-b", body,
		)
		if err != nil {
			return fmt.Errorf("failed to create issue: %s", stdErr.String())
		}

		// Write the new number back to the file
		newURL := strings.TrimSpace(stdOut.String())
		newNumber := filepath.Base(newURL)

		newContent := fmt.Sprintf("---\nnumber: %s\n%s\n---\n\n%s",
			newNumber,
			getRawFrontmatter(data), // Get the original frontmatter content
			body,
		)

		fmt.Printf("Created Issue #%s. Writing number back to %s...\n", newNumber, file)
		if err := ioutil.WriteFile(file, []byte(newContent), 0644); err != nil {
			return err
		}

		// Rename the file to include the new slug
		slug := generateSlug(title)
		newFileName := fmt.Sprintf("%s-%s.md", newNumber, slug)
		newFilePath := filepath.Join(filepath.Dir(file), newFileName)

		if file != newFilePath {
			fmt.Printf("Renaming file to %s\n", newFilePath)
			if err := os.Rename(file, newFilePath); err != nil {
				return fmt.Errorf("failed to rename file %s to %s: %w", file, newFilePath, err)
			}
		}
		return nil

	} else {
		// 6. UPDATE Issue
		title, _ := meta["title"]
		state, _ := meta["state"]

		fmt.Printf("Updating Issue #%s...\n", number)
		_, stdErr, err := gh.Exec("issue", "edit", number,
			"-R", repoFullName,
			"-t", title,
			"-b", body,
		)
		if err != nil {
			return fmt.Errorf("failed to edit issue: %s", stdErr.String())
		}

		// --- ADDED: Check for local file rename on title update ---
		slug := generateSlug(title)
		newFileName := fmt.Sprintf("%s-%s.md", number, slug)
		currentBaseName := filepath.Base(file)

		if currentBaseName != newFileName {
			newFilePath := filepath.Join(filepath.Dir(file), newFileName)
			fmt.Printf("Updating local filename to match new title: %s\n", newFilePath)
			if err := os.Rename(file, newFilePath); err != nil {
				// Log the error but don't stop the whole push
				fmt.Printf("Warning: failed to rename file %s to %s: %v\n", file, newFilePath, err)
			}
		}
		// --- END ADDED ---

		// 7. Handle State Change
		var currentState struct {
			State string `json:"state"`
		}
		client, err := gh.RESTClient(nil)
		if err != nil {
			return err
		}
		err = client.Get(fmt.Sprintf("repos/%s/issues/%s", repoFullName, number), &currentState)
		if err != nil {
			return fmt.Errorf("could not get state for issue #%s: %w", number, err)
		}

		if state != "" && state != currentState.State {
			switch state {
			case "closed":
				fmt.Printf("Closing Issue #%s...\n", number)
				_, stdErr, err = gh.Exec("issue", "close", number, "-R", repoFullName)
				if err != nil {
					return fmt.Errorf("failed to close issue: %s", stdErr.String())
				}
			case "open":
				fmt.Printf("Reopening Issue #%s...\n", number)
				_, stdErr, err = gh.Exec("issue", "reopen", number, "-R", repoFullName)
				if err != nil {
					return fmt.Errorf("failed to reopen issue: %s", stdErr.String())
				}
			}
		}
	}
	return nil
}

// --- Helper Functions for Parsing ---

// --- ADDED ---
// A regex to find unwanted characters in a slug
var nonSlugChars = regexp.MustCompile(`[^a-z0-9\s-]`)

// A regex to collapse multiple hyphens
var multiHyphen = regexp.MustCompile(`-+`)

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = nonSlugChars.ReplaceAllString(slug, "")
	slug = multiHyphen.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// --- END ADDED ---

// parseFrontmatter is a simple, dependency-free YAML parser
func parseFrontmatter(data []byte) map[string]string {
	meta := make(map[string]string)
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return meta // No frontmatter
	}

	lines := strings.Split(parts[1], "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, ":", 2)
		if len(kv) < 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		// Basic quote stripping
		val := strings.Trim(strings.TrimSpace(kv[1]), "\"'")
		meta[key] = val
	}
	return meta
}

// getRawFrontmatter returns just the content between the '---'
func getRawFrontmatter(data []byte) string {
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// parseBody returns everything after the frontmatter
func parseBody(data []byte) string {
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return string(data) // No frontmatter, return all
	}
	return strings.TrimSpace(parts[2])
}

// usage prints the help text
func usage() {
	fmt.Println("A tool to two-way sync GitHub issues with local markdown files.")
	fmt.Println("\nUsage: gh md-issues [pull|push]")
	fmt.Println("  pull   Fetches remote issues and updates local files.")
	fmt.Println("  push   Pushes local file changes to remote issues.")
}
