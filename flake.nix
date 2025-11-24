{
  description = "GitHub CLI extension to generate a daily report (nippou)";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        # Extract version from git tag or use commit hash
        version =
          if self ? shortRev then
            self.shortRev
          else
            "dev";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "gh-nippou";
          inherit version;

          src = ./.;

          subPackages = [ "." ];

          vendorHash = "sha256-GB5c3ZPhjLw9Kn5mNRBNNNesgJ9m2E92GF+Vgbf4gsY=";

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
          ];
        };
      }
    );
}
