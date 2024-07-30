{
  description = "";

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = nixpkgs.legacyPackages.${system}; in
        {
          devShells.default = with pkgs; mkShell {
            # NIX_LD_LIBRARY_PATH = lib.makeLibraryPath [
            #   sqlite
            #   stdenv.cc.cc
            # ];
            # NIX_LD=builtins.readFile "${stdenv.cc}/nix-support/dynamic-linker";
            buildInputs = [gopls delve go sqlite gotests];
            hardeningDisable = [ "all" ];
            shellHook = ''
              echo Welcome to tribler-arr-shim devshell!
              echo To build and run the project:
              # export LD_LIBRARY_PATH=$NIX_LD_LIBRARY_PATH
              echo "go run cmd/main.go server"
            '';
            };
        }
      );
}
