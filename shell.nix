with (import <nixpkgs> {});

mkShell {
   shellHook = ''
     alias e="emacs -nw";
   '';

   buildInputs = [ 
     go_1_13 
     ((emacsPackagesGen emacs).emacsWithPackages (epkgs: 
        (with epkgs.melpaStablePackages; [ magit go-mode nix-mode ivy swiper ]) ++
        (with epkgs.melpaPackages; [ purescript-mode psc-ide ])))
   ];
}
