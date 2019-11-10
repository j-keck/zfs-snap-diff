{ pkgs, purs }:
import (pkgs.fetchFromGitHub {
  owner = "jmackie";
  repo = "purescript-nix";
  rev = "f3f8cb1191e988a87bf045760f2cd95cb51f9561";
  sha256 = "0lpvxlpm3ikbmc8f8azv8nc5l59q9qczzq5ry2ksnk7jgbsny6fj";
  #rev = "a3fa44a0029bef1006300efcb0dcfa6fb23f6589";
  #sha256 = "0pk2wm7h22cmbcyv0dfq91ih3jahg2mlm1s0bajgpwj21igsf0g6";
}) { inherit pkgs purs; }
