module ZSD.Components.TableX where

import Prelude

import Data.Array as A
import Data.Maybe (Maybe(..))
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, makeStateless)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

import ZSD.Utils.Ops (zipWithIndex)

type Idx = Int

type Props a =
  { header :: Array String
  , rows :: Array a
  , mkRow :: a -> Array JSX
  , onRowSelected :: (Tuple Idx a) -> Effect Unit
  , activeIdx :: Maybe Idx
  }

tableX :: forall a. Props a -> JSX
tableX = makeStateless component \props ->

  R.table
      { className: "table table-hover table-sm"
      , children:
        [ R.thead_ [ R.tr_ $ map (R.text >>> A.singleton >>> R.th_) props.header ]
        , R.tbody_ $ flip map (zipWithIndex props.rows) \t@(Tuple idx x) ->
           R.tr
           { className: guard (props.activeIdx == Just idx) " table-active"
           , style: R.css { cursor: "pointer" }
           , onClick: capture_ $ props.onRowSelected t
           , children: map (A.singleton >>> R.td_) $ props.mkRow x
           }
        ]
      }


  where

    component :: Component (Props a)
    component = createComponent "TableX"
