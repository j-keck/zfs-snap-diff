# `zfs-snap-diff`

in this branch i rewrite the whole codebase.

most of the backend rewrite is done. the ui is the old
code. if the backend rewrite is done, i rewrite the ui.

to run the new backend with the old frontend run:

`go run ./cmd/zfs-snap-diff-oldweb <POOL>`


after the rewrite is done, i can add new features.

implemented new features:

  - works now also with 'legacy' mountpoints
