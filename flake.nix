{
  description = "pots: a tailscale pushover webhooker";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      supportedSystems =
        [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {
      overlay = final: prev: {
        pots = self.packages.${prev.system}.pots;
      };
      nixosModule = import ./module.nix;
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          pots = pkgs.buildGoModule {
            pname = "pots";
            version = "v0.0.2";
            src = ./.;
            vendorHash = null;
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.pots);
      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            shellHook = ''
              PS1='\u@\h:\@; '
              echo "Go `${pkgs.go}/bin/go version`"
            '';
            nativeBuildInputs = with pkgs; [ git go gopls go-tools jo jq ];
          };
        });
    };
}

