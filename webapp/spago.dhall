{ name =
    "zfs-snap-diff-webapp"
, dependencies =
    [ "affjax"
    , "console"
    , "effect"
    , "formatters"
    , "js-timers"
    , "now"
    , "numbers"
    , "psci-support"
    , "react-basic"
    , "simple-json"
    , "stringutils"
    , "test-unit"
    , "unfoldable"
    ]
, packages =
    ./packages.dhall
, sources =
    [ "src/**/*.purs", "test/**/*.purs" ]
}
