{ name = "zfs-snap-diff-webapp"
, dependencies =
  [ "aff"
  , "affjax"
  , "arrays"
  , "bifunctors"
  , "console"
  , "control"
  , "datetime"
  , "debug"
  , "effect"
  , "either"
  , "enums"
  , "exceptions"
  , "foldable-traversable"
  , "foreign"
  , "foreign-object"
  , "formatters"
  , "free"
  , "integers"
  , "js-date"
  , "js-timers"
  , "lists"
  , "maybe"
  , "newtype"
  , "now"
  , "numbers"
  , "partial"
  , "prelude"
  , "psci-support"
  , "quickcheck"
  , "react-basic"
  , "react-basic-classic"
  , "react-basic-dom"
  , "react-basic-textf"
  , "record"
  , "refs"
  , "simple-json"
  , "strings"
  , "stringutils"
  , "test-unit"
  , "transformers"
  , "tuples"
  , "unfoldable"
  , "unordered-collections"
  , "unsafe-coerce"
  , "web-dom"
  , "web-file"
  , "web-html"
  ]
, packages = ./packages.dhall
, sources = [ "src/**/*.purs", "test/**/*.purs" ]
}
