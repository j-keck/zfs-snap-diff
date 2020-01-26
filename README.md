# `zfs-snap-diff`

in this branch i rewrite the whole codebase.

  - backend is implemented in go (as before)
  - frontend in purescript (with react-basic)


you need only `go` to build this project.
the frontend code is decoded in `pkg/webapp/bindata.go`.


to run the new buggy, unfinished code, checkout this branch and build it:

  `go build github.com/j-keck/zfs-snap-diff/cmd/zfs-snap-diff`

and run it per

  `./zfs-snap-diff <POOL>`


[Browse Filesystem](doc/browse-filesystem.png)


after the rewrite is done, i can add new features.

implemented new features:

  - works now also with 'legacy' mountpoints
