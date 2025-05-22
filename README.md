# gh-nippou

This is a CLI extension for GitHub CLI (`gh`). This extension helps you generate daily reports (日報 - nippou) based on your GitHub activities.

Based on [masutaka/github-nippou](https://github.com/masutaka/github-nippou/).

## Installation

1.  Ensure you have the GitHub CLI (`gh`) installed. You can find installation instructions [here](https://github.com/cli/cli#installation).
2.  Install this extension:
    ```sh
    gh extension install ryoppippi/gh-nippou
    ```

## Usage

To generate a daily report:

```sh
gh nippou
```

You can also specify a date or a range:

```sh
gh nippou 
gh nippou --since YYYYMMDD --until YYYYMMDD
```

To find more options, you can use:

```sh
gh nippou --help
```

## Features

*   Fetches contributions (commits, pull requests, issues, reviews) from a specified period.
*   Formats the activities into a daily report.

## Motivation

[masutaka/github-nippou](https://github.com/masutaka/github-nippou) is really nice tool, but I have some issues with it.

- We need to add config in `.git/config` file, which makes it hard to manage with public dotfiles.
- I don't want to install the command on Homebrew or other package manager. This tool is only for GitHub, so why don't we use `gh` command?

## License

This project is licensed under the [MIT License](LICENSE) - see the LICENSE file for details.
