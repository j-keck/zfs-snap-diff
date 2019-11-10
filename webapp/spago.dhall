{ name =
    "zfs-snap-diff-webapp"
, dependencies =
    [ "affjax"
    , "console"
    , "effect"
    , "formatters"
    , "js-timers"
    , "numbers"
    , "psci-support"
    , "react-basic"
    , "simple-json"
    , "stringutils"
    , "unfoldable"
    ]
, packages =
    ./packages.dhall
, sources =
    [ "src/**/*.purs", "test/**/*.purs" ]
}
