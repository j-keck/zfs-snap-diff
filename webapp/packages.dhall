let upstream =
      https://github.com/purescript/package-sets/releases/download/psc-0.13.5-20200103/packages.dhall sha256:0a6051982fb4eedb72fbe5ca4282259719b7b9b525a4dda60367f98079132f30

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
          , repo = "https://github.com/lumihq/purescript-react-basic"
          , version = "v13.0.0"
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
          , repo = "https://github.com/slamdata/purescript-affjax.git"
          , version = "v10.0.0"
          }
      ,react-basic-textf =
       { dependencies =
           [ "foreign"
 	   , "maybe"
 	   , "react-basic"
 	   , "unsafe-coerce"
 	   ]
       , repo = "https://github.com/j-keck/purescript-react-basic-textf.git"
       , version = "v0.3.0"
       }
      }

in  upstream // overrides // additions
