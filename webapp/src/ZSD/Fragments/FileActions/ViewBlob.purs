module ZSD.Fragments.FileActions.ViewBlob where

import Prelude
import Effect.Unsafe (unsafePerformEffect)
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, makeStateless)
import React.Basic.DOM as R
import Web.File.Blob (Blob)
import Web.File.Url as Url

type Props
  = { content :: Blob }

viewBlob :: Props -> JSX
viewBlob =
  makeStateless component \props ->
    let
      src = unsafePerformEffect $ Url.createObjectURL props.content
    in
      R.embed
        { src
        , style:
          R.css
            { width: "90%"
            , height: "800px"
            }
        }
  where
  component :: Component Props
  component = createComponent "ViewBlob"
