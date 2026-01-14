{
  description = "Devshell for authlab";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";

    gitignore.url = "github:hercules-ci/gitignore.nix";
    gitignore.inputs.nixpkgs.follows = "nixpkgs";

    devtools.url     = "github:nanoteck137/devtools";
    devtools.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, nixpkgs, flake-utils, gitignore, devtools, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        version = pkgs.lib.strings.fileContents "${self}/version";
        fullVersion = ''${version}-${self.dirtyShortRev or self.shortRev or "dirty"}'';

        backend = pkgs.buildGoModule {
          pname = "authlab";
          version = fullVersion;
          src = ./.;
          subPackages = ["cmd/authlab"];

          ldflags = [
            "-X github.com/nanoteck137/authlab.Version=${version}"
            "-X github.com/nanoteck137/authlab.Commit=${self.dirtyRev or self.rev or "no-commit"}"
          ];

          tags = ["fts5"];

          vendorHash = "sha256-1HbNz94qQg1dRhl6DAdmJWovy1DX+d4FbMg2U3XjqnI=";

          nativeBuildInputs = [ pkgs.makeWrapper ];

          postFixup = ''
            wrapProgram $out/bin/authlab --prefix PATH : ${pkgs.lib.makeBinPath []}
          '';
        };

        frontend = pkgs.buildNpmPackage {
          name = "authlab-web";
          version = fullVersion;

          src = gitignore.lib.gitignoreSource ./web;
          npmDepsHash = "sha256-HmaXCFxbB1Q0JgD11bMlATq9tvrDAP3BzBcGUlOW2L4=";

          PUBLIC_VERSION=version;
          PUBLIC_COMMIT=self.dirtyRev or self.rev or "no-commit";
          PUBLIC_API_ADDRESS="";

          installPhase = ''
            runHook preInstall
            cp -r build $out/
            runHook postInstall
          '';
        };

        tools = devtools.packages.${system};
      in
      {
        packages = {
          default = backend;
          # test = frontend.overrideAttrs(_: {
          #   PUBLIC_API_ADDRESS="http://nanoteck137.net";
          # });
          inherit backend frontend;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            air
            go
            gopls
            nodejs
            tailwindcss_4

            tools.publishVersion
          ];
        };
      }
    ) // {
      nixosModules.backend = import ./nix/backend.nix { inherit self; };
      nixosModules.frontend = import ./nix/frontend.nix { inherit self; };
      nixosModules.default = { ... }: {
        imports = [
          self.nixosModules.backend
          self.nixosModules.frontend
        ];
      };
    };
}
