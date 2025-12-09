{ pkgs, ... }:
{
  projectRootFile = "flake.nix";

  programs = {
    # Nix formatter
    nixfmt.enable = true;

    # Go formatter
    gofmt.enable = true;
  };
}
