{ goos ? "linux" }:
let

  fetchNixpkgs = {rev, sha256}: builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs-channels/archive/${rev}.tar.gz";
    inherit sha256;
  };

  pkgs = import (fetchNixpkgs {
    rev = "8a9807f1941d046f120552b879cf54a94fca4b38";
    sha256 = "0s8gj8b7y1w53ak138f3hw1fvmk40hkpzgww96qrsgf490msk236";
  }) {};

  # nix-prefetch-git https://github.com/justinwoo/easy-purescript-nix
  easy-ps = import (pkgs.fetchFromGitHub {
    owner = "justinwoo";
    repo = "easy-purescript-nix";
    rev = "a09d4ff6a8e4a8a24b26f111c2a39d9ef7fed720";
    sha256 = "1iaid67vf8frsqfnw1vm313d50mdws9qg4bavrhfhmgjhcyqmb52";
  }) { inherit pkgs; };


  buildInputs =
    (with pkgs; [ dhall nodejs utillinux]) ++
    (with pkgs.nodePackages; [ parcel-bundler node2nix ]) ++
    (with easy-ps; [ purs spago spago2nix ]);


  version =
    let lookup-version = pkgs.stdenv.mkDerivation {
          src = pkgs.nix-gitignore.gitignoreSource [ ".gitignore" "/webapp/" ] ./.;
          name = "lookup-version";
          phases = "buildPhase";
          buildPhase = ''
            mkdir -p $out
            ${pkgs.git}/bin/git -C $src describe --always --tags > $out/version
          '';
        };
    in pkgs.lib.removeSuffix "\n" (builtins.readFile "${lookup-version}/version");

  webapp =
    let

      # regenerate spago packages per: (nix-shell --run 'cd webapp; spago2nix generate')
      webapp_ps = (import (./webapp/spago-packages.nix) { inherit pkgs; }).mkBuildProjectOutput {
        # include only directories and purescript source files
        src = with builtins; filterSource (p: t: t == "directory" || match ".*\.purs|.*\.js" p != null) ./webapp;
        purs = easy-ps.purs;
      };

      # regenerate node modules per: (nix-shell; cd webapp; node2nix --nodejs-12)
      webapp_nm = (import ./webapp { inherit pkgs; }).package;

    in pkgs.stdenv.mkDerivation rec {
      inherit version;
      name = "webapp";
      src = pkgs.symlinkJoin {
        name = "webapp-ps+nm";
        paths = [
          "${webapp_ps}"
          "${webapp_nm}/lib/node_modules/webapp"
        ];
      };
      inherit buildInputs;
      phases = "buildPhase";
      buildPhase = ''
        mkdir -p $out
        parcel build --out-dir $out/ ${src}/index.html
      '';
    };


  bindata =
    let
      go-bindata = pkgs.buildGoModule rec {
        name = "go-bindata-${version}";
        version = "8639be0519b3f65dc77b41e3e1df0e54b71fc57e";
        src = pkgs.fetchFromGitHub {
          owner = "go-bindata";
          repo = "go-bindata";
          rev = version;
          sha256 = "11j2cph5w042qx1d91gbwkcq884dlz0lc7ngq1xvyg5hhpd3j8qv";
        };
        modSha256 = "0yc3d1d355l3qirqy7rjb2q8z2r8886nnpfr1714m5i5fcznlj5f";
      };
    in pkgs.runCommand "bindata.go" {} ''
      mkdir -p $out
      cd ${webapp}
      ${go-bindata}/bin/go-bindata -pkg webapp -o $out/bindata.go .
    '';

  zfs-snap-diff = pkgs.buildGoModule rec {
    pname = "zfs-snap-diff";
    inherit version;
    src = pkgs.nix-gitignore.gitignoreSource [ ".gitignore" "/webapp/" ] ./.;
    modSha256 = "1xlgs16lbwdnm2rbzfxwsg5vyc20fsgq506w2ms46a9z3i06zmv1";

    preBuild = ''
      export GOOS=${goos}
      cp -v ${bindata}/bindata.go pkg/webapp/bindata.go
    '';

    CGO_ENABLED = 0;

    buildFlagsArray = ''
      -ldflags=
      -X main.version=${version}
    '';

    installPhase = ''
      mkdir -p $out/bin

      BIN_PATH=${if goos == pkgs.stdenv.buildPlatform.parsed.kernel.name
                 then "$GOPATH/bin"
                 else "$GOPATH/bin/${goos}_$GOARCH"}
      cp $BIN_PATH/zfs-snap-diff $out/bin
    '';
  };
in

if pkgs.lib.inNixShell then pkgs.mkShell {

  buildInputs = buildInputs ++ (with pkgs;
                [ go_1_13
                  ((emacsPackagesGen emacs).emacsWithPackages (epkgs:
                    (with epkgs.melpaStablePackages; [ magit go-mode nix-mode ivy swiper ]) ++
                    (with epkgs.melpaPackages; [ purescript-mode psc-ide ])))
                ]);

  shellHooks = ''
    alias serv="parcel serve --host 0.0.0.0 index.html"
  '';
}
else {
  inherit webapp bindata zfs-snap-diff;
}
