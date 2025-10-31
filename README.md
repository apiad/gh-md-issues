# gh-md-issues

A GitHub CLI extension to two-way sync GitHub issues with local markdown files.

This tool allows you to pull all your repository's issues into a local directory of markdown files, edit them with your favorite text editor, and then push those changes (including new issues, edits, and state changes) back to GitHub.

## Features

  * **Pull Issues:** Fetches all remote issues into `issues/` (open) and `issues/closed/`.
  * **Push Changes:** Pushes local changes back to GitHub.
      * Create new issues by just creating a new file.
      * Update issue titles, bodies, and labels.
      * Close or reopen issues by changing their frontmatter or moving the file.
  * **Slug-based Filenames:** Files are named `NUMBER-SLUG.md` for easy searching (e.g., `123-my-feature-request.md`).
  * **Offline-Friendly:** Make all your changes locally and `push` when you're back online.

## Installation

### Prerequisites

You must have the [GitHub CLI `gh`](https://cli.github.com) installed.

### From Repository (Recommended)

You can install this extension using the `gh extension` command.

```sh
gh extension install apiad/gh-md-issues
```

### From Source

If you have the source code checked out locally:

```sh
# Clone the repository
git clone https://github.com/apiad/gh-md-issues
cd gh-md-issues

# Install dependencies and build
make

# Install the extension from the local path
make install
```

## Usage

The extension has two main commands: `pull` and `push`.

### `gh md-issues pull`

Fetches all issues from your GitHub repository and syncs them to your local `issues/` directories.

```sh
gh md-issues pull
```

  * **Open issues** are saved in `issues/`.
  * **Closed issues** are saved in `issues/closed/`.
  * **Filenames** are created in the format `NUMBER-SLUG.md` based on the issue title.
  * If you re-run `pull`, it will update any existing files and automatically rename them if their titles have changed on GitHub.

### `gh md-issues push`

Detects local changes to your issue files and pushes them to GitHub.

```sh
gh md-issues push
```

The `push` command is driven by `git status` and supports three actions:

**1. Create a New Issue:**

  * Create a new file in the `issues/` directory (e.g., `issues/my-new-idea.md`).
  * Add YAML frontmatter with at least a `title`.
  * Run `gh md-issues push`.
  * The tool will create the issue on GitHub, get the new issue number, and update the file's frontmatter and filename (e.g., `issues/124-my-new-idea.md`).

**2. Update an Existing Issue:**

  * Edit the body or frontmatter (like `title` or `state`) of any issue file.
  * Run `gh md-issues push`.
  * The tool will update the issue on GitHub.
  * If you changed the `title`, the tool will automatically rename your local file to match the new slug.

**3. Close or Reopen an Issue:**

  * **To Close:**
      * **Method 1:** Edit the frontmatter to set `state: closed`.
      * **Method 2 (if using Git):** `git mv issues/124-my-idea.md issues/closed/`
  * **To Reopen:**
      * **Method 1:** Edit the frontmatter to set `state: open`.
      * **Method 2 (if using Git):** `git mv issues/closed/124-my-idea.md issues/`
  * Run `gh md-issues push`. The tool will update the issue's state on GitHub.

## Issue File Format

Local issues are stored as markdown files with YAML frontmatter.

```markdown
---
number: 1
title: "Names with slugs"
state: open
labels:
- enhancement
- help wanted
---

Make markdown filename have the format `NUMBER-SLUG` with slugs from the title of the issue, so we can better search them with `ls`.
```

  * `number`: The GitHub issue number. (Leave blank for new issues)
  * `title`: The issue title. (Required for new issues)
  * `state`: `open` or `closed`.
  * `labels`: A YAML list of labels. (Note: `push` does not currently update labels, this is populated by `pull`.)
  * The content after the `---` is the issue body.

## Contributing

Pull requests are welcome! This project is intended to be a simple, dependency-free tool.

For major changes or new features, please open an issue first to discuss what you would like to change.

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

## License

This project is licensed under the MIT License.
