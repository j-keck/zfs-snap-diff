let upstream =
      https://github.com/purescript/package-sets/releases/download/psc-0.14.3-20210811/packages.dhall sha256:a2de7ef2f2e753733eddfa90573a82da0c7c61d46fa87d015b7f15ef8a6e97d5

let overrides = {=}

let additions =
      { affjax =
        { dependencies =
          [ "aff"
          , "argonaut-core"
          , "arraybuffer-types"
          , "foreign"
          , "form-urlencoded"
          , "http-methods"
          , "integers"
          , "math"
          , "media-types"
          , "nullable"
          , "refs"
          , "unsafe-coerce"
          , "web-xhr"
          ]
        , repo = "https://github.com/slamdata/purescript-affjax.git"
        , version = "v10.0.0"
        }
      , react-basic-textf =
        { dependencies = [ "foreign", "maybe", "react-basic", "unsafe-coerce" ]
        , repo = "https://github.com/j-keck/purescript-react-basic-textf.git"
        , version = "v0.3.0"
        }
      }

in  upstream // overrides // additions
