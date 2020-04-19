with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    python3Packages.selenium
    mypy
    firefox
  ];
}
