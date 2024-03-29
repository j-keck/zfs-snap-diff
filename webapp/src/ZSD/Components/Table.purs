module ZSD.Components.Table where

import Prelude
import Data.Array as A
import Effect (Effect)
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, makeStateless)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

type Props a
  = { header :: Array JSX
    , rows :: Array a
    , mkRow :: a -> Array JSX
    , onRowSelected :: a -> Effect Unit
    }

table :: forall a. Props a -> JSX
table =
  makeStateless component \props ->
    R.table
      { className: "table table-hover table-sm"
      , children:
        [ R.thead_ [ R.tr_ $ map (A.singleton >>> R.th_) props.header ]
        , R.tbody_
            $ flip map props.rows \r ->
                R.tr
                  { style: R.css { cursor: "pointer" }
                  , onClick: capture_ $ props.onRowSelected r
                  , children: map (A.singleton >>> R.td_) $ props.mkRow r
                  }
        ]
      }
  where
  component :: Component (Props a)
  component = createComponent "Table"
