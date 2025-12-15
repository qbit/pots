{
  description = "pots: a tailscale pushover webhooker";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      overlays.default = final: prev: {
        pots = self.packages.${prev.stdenv.hostPlatform.system}.pots;
      };
      nixosModules.default = import ./module.nix;
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          pots = pkgs.buildGoModule {
            pname = "pots";
            version = "v1.0.0";
            src = ./.;
            vendorHash = null;
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            shellHook = ''
              PS1='\u@\h:\@; '
              echo "Go `${pkgs.go}/bin/go version`"
            '';
            nativeBuildInputs = with pkgs; [
              git
              go
              gopls
              go-tools
              jo
              jq
            ];
          };
        }
      );
    };
}
