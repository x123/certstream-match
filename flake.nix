{
  description = "flake for Golang 1.22 devenv";

  inputs.nixpkgs.url = "nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
        };
      });
    in
    {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = [
            # go (version is specified by overlay)
            pkgs.go_1_23

            # goimports, godoc, etc.
            pkgs.gotools

            # https://github.com/golangci/golangci-lint
            pkgs.golangci-lint
          ];
        };
      });
    };
}
