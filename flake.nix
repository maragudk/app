{
  description = "Experimental Nix flake for building and developing the app";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (
        pkgs:
        let
          inherit (pkgs) lib;
        in
        {
          default = pkgs.buildGo127Module {
            pname = "app";
            version = self.shortRev or self.dirtyShortRev or "dev";

            # Everything tracked in git except files that don't affect the build,
            # so editing docs or workflows doesn't trigger a rebuild.
            src = lib.fileset.toSource {
              root = ./.;
              fileset = lib.fileset.difference ./. (
                lib.fileset.unions [
                  ./.github
                  ./docs
                  (lib.fileset.maybeMissing ./flake.lock)
                  ./flake.nix
                ]
              );
            };

            # Update this whenever go.mod or go.sum changes.
            # Set it to lib.fakeHash, run nix build, and copy the hash from the error.
            vendorHash = "sha256-weqxOti+puOgddzr06a5VG/Pl3yL22st2rBtcsWH+vU=";

            subPackages = [ "cmd/app" ];
            tags = [
              "sqlite_fts5"
              "sqlite_math_functions"
            ];
            env.CGO_ENABLED = 1;

            # The tests need an S3 container, and the regular CI runs them.
            doCheck = false;

            nativeBuildInputs = [ pkgs.tailwindcss_4 ];

            # Mirror the layout of the Docker runner image:
            # bin/app, public/ (with styles/app.css), and sqlite/migrations/.
            # Run the app from $out so it finds public/ and sqlite/migrations/.
            postInstall = ''
              mkdir -p $out/sqlite
              cp -r public $out/public
              cp -r sqlite/migrations $out/sqlite/migrations

              # Like the Docker CSS stage, scan only html/ in the working directory.
              # tailwind.css points at glue's html package in the Go module cache,
              # which doesn't exist here, so add the vendored copy as a source.
              mkdir css
              cp tailwind.css css/tailwind.css
              cp -r html css/html
              echo '@source "../vendor/maragu.dev/glue/html";' >> css/tailwind.css
              (cd css && tailwindcss -i tailwind.css -o $out/public/styles/app.css --minify)
            '';

            meta.mainProgram = "app";
          };
        }
      );

      devShells = forAllSystems (
        pkgs:
        let
          # golangci-lint 2.12.2 in nixos-26.05 is built with Go 1.26 and can't lint Go 1.27 code,
          # so build a newer release with Go 1.27, like nixpkgs-unstable does.
          golangci-lint =
            (pkgs.golangci-lint.override { buildGo126Module = pkgs.buildGo127Module; }).overrideAttrs
              (finalAttrs: _: {
                version = "2.15.0";
                src = pkgs.fetchFromGitHub {
                  owner = "golangci";
                  repo = "golangci-lint";
                  tag = "v${finalAttrs.version}";
                  hash = "sha256-m9PoJHescRuT9DiKsb6LL+lQlc7RKAWDiRyf8pNMffA=";
                };
                vendorHash = "sha256-gWbhZc/BJwuiwyTMaR1ShEBI1lo0bGLCXsdj5n91baw=";
                doCheck = false;
              });
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go_1_27
              pkgs.tailwindcss_4
              golangci-lint
              pkgs.gotools
              pkgs.sqlite
            ];
          };
        }
      );
    };
}
