# gh-nippou

This is a CLI extension for the GitHub CLI (`gh`) that helps you generate daily reports (日報 - nippou) based on your GitHub activities.

Based on [masutaka/github-nippou](https://github.com/masutaka/github-nippou).

## Installation

### Using `gh extension`

1.  Ensure you have the GitHub CLI (`gh`) installed. You can find installation instructions [here](https://github.com/cli/cli#installation).
2.  Install this extension:
    ```sh
    gh extension install ryoppippi/gh-nippou
    ```

### Using Nix

You can install `gh-nippou` using Nix flakes.

#### Direct installation

```sh
nix profile install github:ryoppippi/gh-nippou
```

#### In a Nix flake

Add the input to your `flake.nix`:

```nix
{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    gh-nippou.url = "github:ryoppippi/gh-nippou";
  };
}
```

Then use the package directly:

```nix
# In your configuration
environment.systemPackages = [
  inputs.gh-nippou.packages.${system}.default
];
```

#### With Home Manager (using overlay)

Add the input and overlay to your configuration:

```nix
{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    gh-nippou = {
      url = "github:ryoppippi/gh-nippou";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, gh-nippou, ... }: {
    # Apply the overlay to nixpkgs
    nixpkgs.overlays = [ gh-nippou.overlays.default ];
  };
}
```

Then add `gh-nippou` to your Home Manager packages:

```nix
# In your home.nix
home.packages = with pkgs; [
  gh-nippou
];
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

[masutaka/github-nippou](https://github.com/masutaka/github-nippou) is a really nice tool, but I have some issues with it.

- We need to add a config in the `.git/config` file, which makes it hard to manage with public dotfiles.
- I don't want to install the command via Homebrew or another package manager. This tool is only for GitHub, so why not use the `gh` command?

## License

This project is licensed under the [MIT License](LICENSE) - see the LICENSE file for details.

