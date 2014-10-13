#!/usr/bin/env perl
#
# quick and dirty build script
#
use v5.14;
use strict;
use warnings;
use diagnostics;
use IO::Compress::Zip qw(zip $ZipError);

# supported platforms
my %os_arch = (
  freebsd => ["i386", "amd64"],
  linux   => ["amd64"],
  solaris => ["amd64"],
);

# get version from git
my $version = `git describe`;
chomp($version);

# go dependencies
say "get go dependencies";
say `go get -v`;

# bindata.go
say "generate bindata.go ...";
say `sh ./gen-bindata.sh`;


# create build-output
mkdir("build-output");

# build
for my $os(keys(%os_arch)){
    for my $arch(@{$os_arch{$os}}){
        # set build env
        $ENV{"GOOS"} = $os;
        $ENV{"GOARCH"} = ($arch eq "i386" ? "386" : $arch);

        # build it
        my $cmd = qq{go build -ldflags "-X main.VERSION $version" -o zfs-snap-diff};
        say "build for $os/$arch";
        say $cmd;
        system($cmd) && die "build error";

        # pack it
        zip ["zfs-snap-diff", "LICENSE"] => "build-output/zfs-snap-diff-${version}-${os}-${arch}.zip" || die "zip failed: $ZipError\n";

        # delete org binary
        unlink "zfs-snap-diff" || die $!;
    }
}
