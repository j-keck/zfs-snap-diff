{ pkgs ? import <nixpkgs> { }, purs ? "v0.13.0"
, ps-package-sets ? "psc-0.13.0-20190601"
, ps-nix ? import ./purescript-nix.nix { inherit pkgs purs; }, app-src ? ../. }:
let

  app = ps-nix.compile {
    name = "app";
    src = pkgs.lib.cleanSource app-src;
    srcDirs = ["src"];

    dependencies =
    [ "prelude" "console" "react-basic" "affjax" "effect" "simple-json" ];

    package-set = ps-nix.package-sets."${ps-package-sets}";
  };

  # to change npm dependencies:
  #   - edit ./node-modules/node-packages.json
  #   - run (cd node-modules; nix-shell -p nodePackages.node2nix --run 'node2nix -8 -i node-packages.json')
  nodeModules = with (import ./node-modules { });
  pkgs.symlinkJoin {
    name = "node-modules";
    paths = [
      "${react}/lib/node_modules"
      "${react}/lib/node_modules/react/node_modules"
      "${react-dom}/lib/node_modules"
      "${react-dom}/lib/node_modules/react-dom/node_modules"
    ];
  };

in pkgs.runCommand "dashboard" { buildInputs = [pkgs.nodePackages.parcel-bundler]; } ''
  mkdir $out

  ln -s ${nodeModules} node_modules
  cp ${app.src}/index.html .
  cp ${app.src}/app.js .
  mkdir output; cp -va ${app}/* output/
  parcel build --out-dir $out index.html
''
