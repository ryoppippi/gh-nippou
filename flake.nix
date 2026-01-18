{
  description = "GitHub CLI extension to generate a daily report (nippou)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      treefmt-nix,
      git-hooks,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      flake = {
        overlays.default = final: prev: {
          gh-nippou = final.callPackage (
            { buildGoModule }:
            buildGoModule {
              pname = "gh-nippou";
              version = if self ? shortRev then self.shortRev else "dev";

              src = self;

              subPackages = [ "." ];

              vendorHash = "sha256-8eKVjAoeCxZqFZvDcgR2NHhJkNiRKMOQt/6g1Nm+2q4=";

              ldflags = [
                "-s"
                "-w"
                "-X main.version=${if self ? shortRev then self.shortRev else "dev"}"
              ];

              meta = with final.lib; {
                description = "GitHub CLI extension to generate a daily report (nippou)";
                homepage = "https://github.com/ryoppippi/gh-nippou";
                license = licenses.mit;
                maintainers = with maintainers; [ ryoppippi ];
              };
            }
          ) { };
        };
      };

      perSystem =
        { pkgs, system, ... }:
        let
          # Extract version from git tag or use commit hash
          version = if self ? shortRev then self.shortRev else "dev";

          treefmtEval = treefmt-nix.lib.evalModule pkgs ./treefmt.nix;

          pre-commit-check = git-hooks.lib.${system}.run {
            src = ./.;
            hooks = {
              treefmt = {
                enable = true;
                package = treefmtEval.config.build.wrapper;
              };
            };
          };
        in
        {
          formatter = treefmtEval.config.build.wrapper;

          checks = {
            formatting = treefmtEval.config.build.check self;
            inherit pre-commit-check;
          };

          packages.default = pkgs.buildGoModule {
            pname = "gh-nippou";
            inherit version;

            src = ./.;

            subPackages = [ "." ];

            vendorHash = "sha256-8eKVjAoeCxZqFZvDcgR2NHhJkNiRKMOQt/6g1Nm+2q4=";

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
            ];

            meta = with pkgs.lib; {
              description = "GitHub CLI extension to generate a daily report (nippou)";
              homepage = "https://github.com/ryoppippi/gh-nippou";
              license = licenses.mit;
              maintainers = with maintainers; [ ryoppippi ];
            };
          };

          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gotools
              go-tools
              just
            ];
            inherit (pre-commit-check) shellHook;
          };
        };
    };
}
