#!/bin/sh
# set -e
#
# this is a temporary script to build the alpha versions

version=$(git describe --always)
for goos in linux freebsd darwin; do
    echo "==============="
    echo "build for $goos"
    build=$(nix-build -A zfs-snap-diff --argstr goos $goos)
    cp -v $build/bin/zfs-snap-diff zfs-snap-diff-$version
    cp -v $build/bin/zsd zsd-$version
    tar cvfz zfs-snap-diff-$goos-$version.tgz zfs-snap-diff-$version zsd-$version

    rm -f zfs-snap-diff-$version zsd-$version
done
