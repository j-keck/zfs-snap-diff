let upstream =
      https://github.com/purescript/package-sets/releases/download/psc-0.13.4-20191025/packages.dhall sha256:f9eb600e5c2a439c3ac9543b1f36590696342baedab2d54ae0aa03c9447ce7d4

let overrides =
      { react-basic =
          { dependencies =
              [ "exceptions"
              , "effect"
              , "console"
              , "web-events"
              , "web-html"
              , "foreign-object"
              , "aff"
              , "unsafe-coerce"
              , "record"
              , "web-dom"
              , "nullable"
              , "functions"
              ]
          , repo =
              "https://github.com/lumihq/purescript-react-basic"
          , version =
              "v13.0.0"
          }
      }

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
          , repo =
              "https://github.com/slamdata/purescript-affjax.git"
          , version =
              "v10.0.0"
          }
      }

in  upstream ⫽ overrides ⫽ additions