

{ pkgs ? import <nixpkgs> {} }:
let

  easy-ps = import (pkgs.fetchFromGitHub {
    owner = "justinwoo";
    repo = "easy-purescript-nix";
    rev = "aa94aeac3a6ad9b4dfa0e807ad1421097d74f663";
    sha256 = "1kfhi6rscgf165zg4f1s0fgppygisvc7dppxb93n02rypxfxjirm";
  }) { inherit pkgs; };


  # to change npm dependencies:
  #   - edit ./nix/node-modules/node-packages.json
  #   - run (cd nix/node-modules; nix-shell -p nodePackages.node2nix --run 'node2nix --nodejs-10 -i node-packages.json')
  nodeModules = with (import ./nix/node-modules { });
    pkgs.symlinkJoin {
      name = "node-modules";
      paths = [
        "${react}/lib/node_modules"
        "${react}/lib/node_modules/react/node_modules"
        "${react-dom}/lib/node_modules"
        "${react-dom}/lib/node_modules/react-dom/node_modules"
        "${bootstrap}"
        "${octicons}"
      ];
    };

  ignoreSource = [
    "default.nix"
    "nix/"
  ];

  # to regenerate './nix/spago-packages.nix' run: nix-shell --run 'spago2nix generate && mv spago-packages.nix nix/'
  app = (import (./nix/spago-packages.nix) { inherit pkgs; }).mkBuildProjectOutput {
    src = pkgs.nix-gitignore.gitignoreSource ignoreSource ./.;
    purs = easy-ps.purs;
  };


  buildInputs =
    (with pkgs; [ dhall nodejs ]) ++
    (with pkgs.nodePackages; [ parcel-bundler ]) ++
    (with easy-ps; [ purs spago spago2nix ]);


in
if pkgs.lib.inNixShell then pkgs.mkShell {
  inherit buildInputs;
  shellHooks = ''
    alias serv="parcel serve --host 0.0.0.0 index.html"
  '';
}
else pkgs.stdenv.mkDerivation rec {
  name = "zsd-webapp";
  version = "0.1.0";
  src = pkgs.nix-gitignore.gitignoreSource ignoreSource ./.;

  inherit buildInputs;

  phases = "buildPhase";

  buildPhase = ''
    mkdir -p $out/webapp

    ln -s ${nodeModules} $out/webapp/node_modules
    ln -s ${app}/output $out/webapp/output

    cp ${src}/index.html $out/webapp/
    cp ${src}/webapp.js $out/webapp/

    #parcel build --out-dir $out/webapp index.html
  '';
}
