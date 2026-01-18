{ pkgs, ... }:
{
  projectRootFile = "flake.nix";

  programs = {
    # Nix formatter
    nixfmt.enable = true;
    nixfmt.package = pkgs.nixfmt;

    # Go formatter
    gofmt.enable = true;
  };
}
