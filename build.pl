#!/usr/bin/env perl
#
# build script for 'zfs-snap-diff'
#
#
use v5.20;
use strict;
use warnings;
use diagnostics;
use IO::Compress::Zip qw(zip $ZipError);
use experimental qw{smartmatch};

# supported platforms
my %os_arch = (
    freebsd => ["i386", "amd64"],
    linux   => ["amd64"],
    solaris => ["amd64"],
    );

# first argument is the programm mode
# set to empty string if no mode are given to prevent 'Use of uninitialized value $mode...' warning
my $mode = shift;
$mode = "" if(!defined($mode));

# action
for($mode) {
    &build when "build";
    &webdev when "webdev";
    &release when "release";
    &help when ["-h", "help"];
    default {
        say "invalid mode: '$mode'";
        &help;
    }
}

#
# print the help
#
sub help {
    say <<"EOF";

usage: $0 <MODE>

where <MODE> can be:
  build:   build the binary for the actual platform
  webdev:  start 'zfs-snap-diff' with the option to serve the frontend code from the 'webapp' directory
  release: build and create a zip for each supported platform under 'build-output/'
  help:    show this help
EOF
}


#
# build the executable
#
sub build {

    # get the version from git
    my $version = &git_describe();

    # generate bindata.go
    &gen_bindata();

    # build it
    my $cmd = qq{go build -ldflags "-X main.VERSION=$version" -o zfs-snap-diff};
    say "build 'zfs-snap-diff' ($cmd)";
    system($cmd) == 0 or die "build error";
}


#
# build the executable for each supported platform and create zip under 'build-output/'
#
sub release {

    # get the version from git
    my $version = &git_describe();

    # create build-output
    mkdir("build-output");

    # build for every platform
    for my $os(keys(%os_arch)){
        for my $arch(@{$os_arch{$os}}){
            # set build env
            $ENV{"GOOS"} = $os;
            $ENV{"GOARCH"} = ($arch eq "i386" ? "386" : $arch);

            say "build for $os/$arch";
            &build();

            # pack it
            zip ["zfs-snap-diff", "LICENSE"] => "build-output/zfs-snap-diff-${version}-${os}-${arch}.zip" or die "zip failed: $ZipError\n";

            # delete the binary
            unlink "zfs-snap-diff" || die $!;
        }
    }
}


#
# start 'zfs-snap-diff' with the option to serve the frontend code from the 'webapp' directory
#
sub webdev {
    my $cmd = "ZSD_SERVE_FROM_WEBAPP=YES ./zfs-snap-diff @ARGV";
    say "exec '$cmd'";
    exec($cmd);
}


#
# get the actual version
#
sub git_describe {
    # validate that 'git' is installed
    system("git version 2>&1 > /dev/null") == 0 or
        die "'git' missing!";

    # get version from git
    chomp(my $version = qx{git describe});
    return $version;
}


#
# generate 'bindata.go' per 'go-bindata' cmd
#
sub gen_bindata {
    # validate that 'go-bindata' is installed
    system("go-bindata -version 2>&1 > /dev/null") == 0 or
        die "'go-bindata' missing! - please install per: 'go get -u github.com/jteeuwen/go-bindata/...'";

    my @ignore = qw{go-bindata .git config.json README angular-mocks.js 'emacs.*core'};
    my $cmd = "go-bindata " . join(" ", map("-ignore=$_", @ignore)) . " webapp/...";

    say "generate 'bindata.go': ($cmd)";
    system($cmd) == 0 or die "unable to build 'bindata.go'";
}
