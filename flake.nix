{
  description = "A quick healthcheck for your go code";

  inputs.nixpkgs.url = "nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      version = builtins.substring 0 8 self.lastModifiedDate;
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      
      nixpkgsFor = forAllSystems (system: import nixpkgs {
        inherit system;
        overlays = [
          (final: prev: { go = prev.go_1_24; })
        ];
      });
    in
    {
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in
        rec {
          checkup = pkgs.buildGoModule {
            pname = "checkup";
            inherit version;
            src = ./.;
            vendorHash = "sha256-kWAkezWJ/FUTQrVeFPT8Irr5obe4zIVtsJ08ql6LCmI=";
          };
          default = checkup;
        }
      );

      apps = forAllSystems (system: rec {
        checkup = {
          type = "app";
          program = "${self.packages.${system}.checkup}/bin/checkup}";
        };
        default = checkup;
      });

      defaultPackages = forAllSystems (system: self.packages.${system}.default);
      defaultApp = forAllSystems (system: self.apps.${system}.default);

      devshells.default = forAllSystems (system: 
        let pkgs = nixpkgsFor.${system};
        in with pkgs; mkShell {
          buildInputs = [ go_1_24 gotools go-tools gopls nixpkgs-fmt ];
        }
      );
    };

}
