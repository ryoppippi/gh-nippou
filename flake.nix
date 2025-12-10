{
  description = "GitHub CLI extension to generate a daily report (nippou)";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
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
    {
      self,
      nixpkgs,
      flake-utils,
      treefmt-nix,
      git-hooks,
    }:
    {
      overlays.default = final: prev: {
        gh-nippou = final.callPackage (
          { buildGoModule }:
          buildGoModule {
            pname = "gh-nippou";
            version = if self ? shortRev then self.shortRev else "dev";

            src = self;

            subPackages = [ "." ];

            vendorHash = "sha256-dvidDJ6NT+naeYCi/I2KykqwcUIPARvKt82ScakQNLQ=";

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${if self ? shortRev then self.shortRev else "dev"}"
            ];

            meta = with final.lib; {
              description = "GitHub CLI extension to generate a daily report (nippou)";
              homepage = "https://github.com/ryoppippi/gh-nippou";
              license = licenses.mit;
              maintainers = [ ];
            };
          }
        ) { };
      };
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
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

          vendorHash = "sha256-dvidDJ6NT+naeYCi/I2KykqwcUIPARvKt82ScakQNLQ=";

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];

          meta = with pkgs.lib; {
            description = "GitHub CLI extension to generate a daily report (nippou)";
            homepage = "https://github.com/ryoppippi/gh-nippou";
            license = licenses.mit;
            maintainers = [ ];
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
      }
    );
}
