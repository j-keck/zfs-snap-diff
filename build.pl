#!/usr/bin/env perl
#
# quick and dirty build script
#
use v5.14;
use strict;
use warnings;
use diagnostics;
use IO::Compress::Zip qw(zip $ZipError);

# get version from git
my $version = `git describe`;
chomp($version);

# bindata.go
say "generate bindata.go ...";
say `go-bindata -ignore=.git -ignore=config.json -ignore=README webapp/...`;


# create build-output
mkdir("build-output");

# build
for my $os(<freebsd linux>){
    for my $arch(<i386 amd64>){
        # set build env
        $ENV{"GOOS"} = $os;
        $ENV{"GOARCH"} = ($arch eq "i386" ? "386" : $arch);

        # build it
        my $cmd = qq{go build -ldflags "-X main.VERSION $version" -o zfs-snap-diff};
        say "build for $os/$arch";
        say $cmd;
        system($cmd) && die "build error";

        # pack it
        zip "zfs-snap-diff" => "build-output/zfs-snap-diff-${version}-${os}-${arch}.zip" || die "zip failed: $ZipError\n";

        # delete org binary
        unlink "zfs-snap-diff" || die $!;
    }
}
