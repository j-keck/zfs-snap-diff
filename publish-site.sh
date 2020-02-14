#!/bin/sh
set -e
#
# this is a temporary script to build and deploy
# the site (in the gh-pages branch).
#
# when the new version hit's the master branch,
# this script will be replaced with a github action
#


# ################################################
echo "cleanup / setup"
rm -rf gh-pages
mkdir gh-pages


# ################################################
echo "build site"
cp -r $(nix-build --no-out-link --no-build-output -A site)/. gh-pages/
chmod -R u+w gh-pages/


# ################################################
echo "build binaries"
VERSION=$(git describe --always)
for GOOS in linux freebsd darwin solaris; do
    echo "  for $GOOS"
    BUILD=$(nix-build --no-out-link --no-build-output -A zfs-snap-diff --argstr goos $GOOS)
    cp -fv $BUILD/bin/zfs-snap-diff zfs-snap-diff-$VERSION
    cp -fv $BUILD/bin/zsd zsd-$VERSION

    ARCHIVE=zfs-snap-diff-$GOOS-$VERSION.tgz
    tar cvfz gh-pages/$ARCHIVE zfs-snap-diff-$VERSION zsd-$VERSION

    rm -fv zfs-snap-diff-$version zsd-$version
done


# ################################################
echo "publish"
cd gh-pages
git init
git add -A
git -c user.name='Jürgen Keck' -c user.email='jhyphenkeck@gmail.com' commit -m 'regen gh-pages'
git push -f -q git@github.com:j-keck/zfs-snap-diff.git HEAD:gh-pages
