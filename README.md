# `zfs-snap-diff`: compare / restore files from zfs snapshots
Next features of `zfs-snap-diff`: [feature poll](https://github.com/j-keck/zfs-snap-diff/issues?q=is%3Aissue+is%3Aopen+label%3Afeature-poll). Please comment / vote.

## Description

With `zfs-snap-diff` you can explore file differences and restore changes from older file versions in different zfs snapshots.
You can restore the whole file from a older version, or select single changes to revert in the 'Diff' view.

  
`zfs-snap-diff` has a web frontend, so it can run on your local work machine or on your remote file / backup server (no Xserver necesarry).

To keep it protable and independent, it's made as a single executable with all html / js stuff included.
The backend is implemented in golang, the frontend with [angularjs](https://angularjs.org), [bootstrap](http://getbootstrap.com) and [highlight.js](https://github.com/isagalaev/highlight.js).


  
##Usage
_under linux, you need the '-use-sudo' flag if you don't run it as root - see the options below_

### Startup a server instance

      ./zfs-snap-diff [OPT_ARGS] <ZFS_NAME>
  
  * starts a web server on port http://127.0.0.1:12345
  * optional arguments:
    * -a: listen on all interfaces (default: listen only on localhost)
    * -p: web server port (default: 12345)
    * -default-file-action: file action when a file is selected (default: view):
      * off: no action
      * view: view the file from the given snapshot
      * diff: diff the file from the given snapshot with the actual version
      * download: download the file from the given snapshot
      * restore: restore the file from the given snapshot
    * -diff-context-size: context size in diff (default: 5)
    * -scan-snap-limit: limit how many snapshots are scan to search older file version (default: scan all)
      * negative limit: scan all snapshots
      * recommended if you have many snapshots
    * -compare-file-method: compare method when searching in snapshots for other file versions (default: auto)
      * supported methods:
        * auto: compares text files per md5, others by size+modTime
        * size+modTime: compares per size and modification time (very cheap)
        * size: compares per size (very cheap)
        * md5: compares per md5 (VERY EXPENSIVE! combine it with '-scan-snap-limit' and use it only for text files!)
    * -use-sudo: use sudo when executing os commands
      * necessary under linux when running as non root
      * adjust sudo rules (see [doc/etc/sudoers.d/zfs-snap-diff](https://github.com/j-keck/zfs-snap-diff/blob/master/doc/etc/sudoers.d/zfs-snap-diff))


  

### Connect with your web browser

      http://localhost:12345



## User guide

### Browse actual filesystem state 

#### Select a dataset

Select a dataset which you would explore. If you start `zfs-snap-diff` on a dataset with no childrens, the current dataset are selected.

![Datasets](doc/zsd-ba-datasets.png)

  
#### Search a file
  
Search a file to compare in the file browser.
    
![File browser](doc/zsd-ba-file-browser.png)


   
#### Select a file

When a file is selected, `zsd-snap-diff` search all snapshots where the selected file was modified (it compares text files per md5, others per size+modTime).
    
![File selected](doc/zsd-ba-snapshots.png)
  

#### Select a snapshot

When you select a snapshot, you can view, diff, download or restore the file from the selected snapshot.

###### View
View the file content from an older file version.
![File View](doc/zsd-ba-view-file.png)

###### Diff
Explore file differences and pick single changes to revert.

intext diff:  
![intext diff](doc/zsd-ba-diff-intext.png)

  
side by side diff:
![side by side diff](doc/zsd-ba-diff-side-by-side.png)



### Browse snapshot state

#### Search a snaphot

Search a snapshot in the snapshot browser. All snapshots from the selected dataset are displayed in this view.
  
![Snapshot Browser](doc/zsd-bs-snapshots.png)


#### Select a snapshot

When a snapshot is selected, the file-browser shows the content from this snapshot.

![File Browser](doc/zsd-bs-file-browser.png)

  
From here you can easy view / restore a deleted file.
  
![File View](doc/zsd-bs-file-selected.png)


 


## Installation
  
### Prebuild

  Get a package for your platform from: https://github.com/j-keck/zfs-snap-diff/releases/latest

 *ping me if your platform is missing*
    
### Manual build

  * clone the repository

        git clone github.com/j-keck/zfs-snap-diff

  * change into the project directory

        cd zfs-snap-diff

  * build

        ./build.pl build

  * run it

        ./zfs-snap-diff <ZFS_NAME>


##Notes

  * if you download a file from a snapshot, the generated file name has the snapshot name included:

        <ORG_FILE_NAME>-<SNAPSHOT_NAME>.<FILE_SUFFIX>

  * if you restore / patch a file, the orginal file will be saved under:

        <ORG_FILE_PATH>/.zsd/<ORG_FLILE_NAME>_<TIMESTAMP>

  * for snapshot differences (Browse snapshot diff), you need to set the diff permission:

        zfs allow -u <USER_NAME> diff <ZFS_NAME>

   

## Coding Notes

  * if you change something under 'webapp/' 

    * start `zfs-snap-diff` per `./build.pl webdev <ZFS_NAME>`
      to serve the static content from the `webapp` folder

    * re-generate bindata.go and recompile `zfs-snap-diff`
      * ./build.pl build


## Changelog

###0.0.X###

0.0.9:
  * show file size and modify timestamp in the file-browser
  * list directories at first in the file-browser
  * sortable columns in the file-browser
  * only regular files / directories are clickable

[all commits from 0.0.8...0.0.9](https://github.com/j-keck/zfs-snap-diff/compare/0.0.8...0.0.9)

0.0.8:
  * dataset selectable in 'browse-actual' view
  * add size informations to dataset (to match 'zfs list' output)
  * small fixes
  * code cleanup
  
[all commits from 0.0.7...0.0.8](https://github.com/j-keck/zfs-snap-diff/compare/0.0.7...0.0.8)
  
0.0.7:
  * support sub zfs filesystems (datasets)
  * optional use sudo when execute zfs commands
    * necessary under linux when running as non root
    * needs sudo rules (see [doc/etc/sudoers.d/zfs-snap-diff](https://github.com/j-keck/zfs-snap-diff/blob/master/doc/etc/sudoers.d/zfs-snap-diff))
    * start `zfs-snap-diff` with '-use-sudo'
  * new view for server messages
  
[all commits from 0.0.6...0.0.7](https://github.com/j-keck/zfs-snap-diff/compare/0.0.6...0.0.7)
  
0.0.6:
  * check if file in snapshot has changed filetype depend:
    * text files: md5
    * others: size+modTime
  * diffs created in the backend (per [go-diff](https://github.com/sergi/go-diff))
    * different presentation: intext / side by side
    * possibility to revert single changes
  
[all commits from 0.0.5...0.0.6](https://github.com/j-keck/zfs-snap-diff/compare/0.0.5...0.0.6)  
   
  
0.0.5:
  * file compare method configurable: size+modTime (default) or md5
  * optional limit how many snapshots are scan to search older file version
  * autohide notifications in frontend
  * show message if no snapshots found
  
[all commits from 0.0.4...0.0.5](https://github.com/j-keck/zfs-snap-diff/compare/0.0.4...0.0.5)  
  
0.0.4:
  * view, diff, download or restore file from a snapshot
  * view file with syntax highlight
  * browse old snapshot versions
  * easy switch "versions" per 'Older' / 'Newer' buttons
  * cleanup frontend
  * refactor backend
  
[all commits 0.0.3...0.0.4](https://github.com/j-keck/zfs-snap-diff/compare/0.0.3...0.0.4)    
  
0.0.3:
  * show server errors on frontend
  * show waiting spinner when loading

[all commits 0.0.2...0.0.3](https://github.com/j-keck/zfs-snap-diff/compare/0.0.2...0.0.3)        
  
0.0.2 :
  * partial frontend configuration from server
  * fix firefox ui

[all commits 0.0.1...0.0.2](https://github.com/j-keck/zfs-snap-diff/compare/0.0.1...0.0.2)      

0.0.1:
  * prototype  
