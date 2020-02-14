+++
title = "Changelog"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 40
+++

## 1.0.0-alpha (unreleased) {#1-dot-0-dot-0-alpha--unreleased}

{{< hint note >}}
This version is a complete rewrite.

The backend is implemented in [Go](https://golang.org) (as before) and the frontend in [PureScript](http://purescript.org).
{{< /hint >}}

-   download a complete directory as zip-archive

-   date-range based search for file versions
    -   this speeds up the scan dramatically if
        there are thousands snapshots on spinning disk

-   bookmark support
    -   bookmarks are per dataset and stored in the browser ([Web storage](https://en.wikipedia.org/wiki/Web%5Fstorage)).

-   works now also with 'legacy' mountpoints

-   new backend and frontend code

[all commits from 0.0.10...dev](https://github.com/j-keck/zfs-snap-diff/compare/0.0.10...dev)


## 0.0.10 {#0-dot-0-dot-10}

-   use relative url for service endpoints
    -   to use zfs-snap-diff behind a reverse proxy
    -   minimal example config snipped for nginx:

        location _zfs-snap-diff_ {
            proxy\_pass <http://localhost:12345/>;
        }

-   optional tls encryption
-   listen address per '-l' flag configurable

[all commits from 0.0.9...0.0.10](https://github.com/j-keck/zfs-snap-diff/compare/0.0.9...0.0.10)


## 0.0.9 {#0-dot-0-dot-9}

-   show file size and modify timestamp in the file-browser
-   list directories at first in the file-browser
-   sortable columns in the file-browser
-   only regular files / directories are clickable

[all commits from 0.0.8...0.0.9](https://github.com/j-keck/zfs-snap-diff/compare/0.0.8...0.0.9)


## 0.0.8 {#0-dot-0-dot-8}

-   dataset selectable in 'browse-actual' view
-   add size informations to dataset (to match 'zfs list' output)
-   small fixes
-   code cleanup

[all commits from 0.0.7...0.0.8](https://github.com/j-keck/zfs-snap-diff/compare/0.0.7...0.0.8)


## 0.0.7 {#0-dot-0-dot-7}

-   support sub zfs filesystems (datasets)
-   optional use sudo when execute zfs commands
    -   necessary under linux when running as non root
    -   needs sudo rules
    -   start \`zfs-snap-diff\` with-'-use-sudo'
-   new view for server messages

[all commits from 0.0.6...0.0.7](https://github.com/j-keck/zfs-snap-diff/compare/0.0.6...0.0.7)


## 0.0.6 {#0-dot-0-dot-6}

-   check if file in snapshot has changed filetype depend:
    -   text files: md5
    -   others: size+modTime
-   diffs created in the backend (per [go-diff](https://github.com/sergi/go-diff))
    -   different presentation: intext / sid- by side
    -   possibility to revert single changes

[all commits from 0.0.5...0.0.6](https://github.com/j-keck/zfs-snap-diff/compare/0.0.5...0.0.6)


## 0.0.5 {#0-dot-0-dot-5}

-   file compare method configurable: size+modTime (default) or md5
-   optional limit how many snapshots are scan to search older file version
-   autohide notifications in frontend
-   show message if no snapshots found

[all commits from 0.0.4...0.0.5](https://github.com/j-keck/zfs-snap-diff/compare/0.0.4...0.0.5)


## 0.0.4 {#0-dot-0-dot-4}

-   view, diff, download or restore file from a snapshot
-   view file with syntax highlight
-   browse old snapshot versions
-   easy switch "versions" per 'Older' / 'Newer' buttons
-   cleanup frontend
-   refactor backend

[all commits 0.0.3...0.0.4](https://github.com/j-keck/zfs-snap-diff/compare/0.0.3...0.0.4)


## 0.0.3 {#0-dot-0-dot-3}

-   show server errors on frontend
-   show waiting spinner when loading

[all commits 0.0.2...0.0.3](https://github.com/j-keck/zfs-snap-diff/compare/0.0.2...0.0.3)


## 0.0.2 {#0-dot-0-dot-2}

-   partial frontend configuration from server
-   fix firefox ui

[all commits 0.0.1...0.0.2](https://github.com/j-keck/zfs-snap-diff/compare/0.0.1...0.0.2)


## 0.0.1 {#0-dot-0-dot-1}

-   prototype
