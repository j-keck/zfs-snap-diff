{ name = "zfs-snap-diff-webapp"
, dependencies =
    [ "affjax"
    , "console"
    , "debug"
    , "effect"
    , "formatters"
    , "js-timers"
    , "now"
    , "numbers"
    , "psci-support"
    , "react-basic"
    , "react-basic-textf"
    , "simple-json"
    , "stringutils"
    , "test-unit"
    , "unfoldable"
    , "unordered-collections"
    ]
, packages = ./packages.dhall
, sources = [ "src/**/*.purs", "test/**/*.purs" ]
}
