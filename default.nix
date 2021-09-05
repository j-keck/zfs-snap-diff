{ goos ? "linux", with-dev-tools ? false }:
let

  pkgs = import (builtins.fetchGit {
    name = "nixos-21.05";
    url = "https://github.com/nixos/nixpkgs/";
    ref = "refs/heads/nixos-21.05";
    rev = "a1007637cea374bd1bafd754cfd5388894c49129";
  }) {};

  # nix-prefetch-git https://github.com/justinwoo/easy-purescript-nix
  easy-ps = import (pkgs.fetchFromGitHub {
    owner = "justinwoo";
    repo = "easy-purescript-nix";
    rev = "5e66c8fe92e80c054cd6ef7e9ac0e91de81175ca";
    sha256 = "1wr5gynay76623mnf0jz8adwvldk4qyqc96r3yp9qkql83gn3zpx";
  }) { inherit pkgs; };


  buildInputs =
    (with pkgs; [ hugo dhall nodejs utillinux]) ++
    (with pkgs.nodePackages; [ parcel-bundler node2nix ]) ++
    (with easy-ps; [ purs spago spago2nix purty ]);


  version =
    let lookup-version = pkgs.stdenv.mkDerivation {
          src = builtins.path { name = "git"; path = ./.git; };
          name = "lookup-version";
          phases = "buildPhase";
          buildPhase = ''
            mkdir -p $out
            ${pkgs.git}/bin/git --git-dir=$src describe --always --tags > $out/version
          '';
        };
    in pkgs.lib.removeSuffix "\n" (builtins.readFile "${lookup-version}/version");

  webapp =
    let

      # regenerate spago packages per: nix-shell --run 'cd webapp; spago2nix generate'
      webapp_ps = (import (./webapp/spago-packages.nix) { inherit pkgs; }).mkBuildProjectOutput {
        src = pkgs.nix-gitignore.gitignoreSourcePure
                [ "/.cache/" "/.spago/" "/node_modules" "/output/" "/dist/" ".psci_modules"] ./webapp;
        purs = easy-ps.purs;
      };

      # regenerate node modules per: nix-shell  --run 'cd webapp; node2nix --nodejs-12'
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
        name = "go-bindata";
        version = "8639be0519b3f65dc77b41e3e1df0e54b71fc57e";
        src = pkgs.fetchFromGitHub {
          owner = "go-bindata";
          repo = "go-bindata";
          rev = version;
          sha256 = "11j2cph5w042qx1d91gbwkcq884dlz0lc7ngq1xvyg5hhpd3j8qv";
        };
        vendorSha256 = "0sjjj9z1dhilhpc8pq4154czrb79z9cm044jvn75kxcjv6v5l2m5";
        modSha256 = "00zr3kpaywqi5kgjzvmf284njxl1fs1k9xaz1b8azwxnjpy77i0c";
      };
    in pkgs.runCommand "bindata.go" {} ''
      mkdir -p $out
      cd ${webapp}
      ${go-bindata}/bin/go-bindata -pkg webapp -o $out/bindata.go .
    '';

  zfs-snap-diff = pkgs.buildGo115Module rec {
    pname = "zfs-snap-diff";
    inherit version;
    src = pkgs.nix-gitignore.gitignoreSource [ ".gitignore" "/webapp/" ] ./.;
    vendorSha256 = "1pr4xnm412ihmvxm3zygqsb34wabyxvs7dlnhbks3sxr0zsfp6fi";
    modSha256 = "0k1sz9mnz09pgn4w3k2dx0grcb66xd3h0f6ccc2r76vz6mz1hpgf";

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
      mkdir -p $out

      BIN_PATH=${if goos == pkgs.stdenv.buildPlatform.parsed.kernel.name
                 then "$GOPATH/bin"
                 else "$GOPATH/bin/${goos}_$GOARCH"}

      mkdir -p $out/bin
      cp $BIN_PATH/zfs-snap-diff $out/bin
      cp $BIN_PATH/zsd $out/bin

      mkdir -p $out/share
      cp LICENSE $out/share
    '';
  };

  site =
    let theme = pkgs.fetchFromGitHub {
          owner = "alex-shpak";
          repo = "hugo-book";
          rev = "dae803fa442973561821a44b08e3a964614d07df";
          sha256 = "0dpb860kddclsqnr4ls356jn4d1l8ymw5rs9wfz2xq4kkrgls4dl";
        };
    in pkgs.stdenv.mkDerivation rec {
         name = "zfs-snap-diff-site";
         inherit version;
         src = ./doc/site;
         buildPhase = ''
          cp -a ${theme}/. themes/book
           ${pkgs.hugo}/bin/hugo --minify
         '';
         installPhase = ''
           cp -r public $out
         '';
       };

in

if pkgs.lib.inNixShell then pkgs.mkShell {

  buildInputs = buildInputs ++
    (if with-dev-tools
       then [ pkgs.go_1_12 pkgs.ispell
                ((pkgs.emacsPackagesGen pkgs.emacs).emacsWithPackages (epkgs:
                    (with epkgs.melpaStablePackages; [ magit go-mode nix-mode ivy swiper ]) ++
                    (with epkgs.melpaPackages; [ purescript-mode psc-ide ])))
            ]
       else []);

  shellHooks = ''
    unset GOPATH
    alias serv="parcel serve --host 0.0.0.0 index.html"

    alias site="hugo --port 54321 --source doc/site server"
  '';
}
else {
  inherit webapp bindata zfs-snap-diff site;
}
