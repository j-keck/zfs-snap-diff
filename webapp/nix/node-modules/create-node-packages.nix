{ pkgs ? import <nixpkgs> {}}:
let
  nodePackagesJson = pkgs.writeText "node-packages.json" ''
    [ "react", "react-dom" ]
  '';

  nodePackagesNix = pkgs.runCommand "x" {} ''
    mkdir $out
    "${pkgs.nodePackages.node2nix}/bin/node2nix" -i ${nodePackagesJson}
  '';
in nodePackagesNix

# in pkgs.symlinkJoin {
#   name = "blub";
#   paths = [
#     nodePackagesJson
#     nodePackagesNix
#   ];
# }
