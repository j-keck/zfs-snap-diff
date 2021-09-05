module ZSD.Components.DropDownButton where

import Prelude
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Foreign.Object as O
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, makeStateless)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

type Props
  = { content :: JSX
    , title :: String
    , disabled :: Boolean
    , onClick :: Effect Unit
    , entries :: Array (Tuple JSX (Effect Unit))
    , entriesTitle :: String
    }

dropDownButton :: Props -> JSX
dropDownButton =
  makeStateless component \props ->
    R.span
      { className: "dropdown"
      , children:
        [ R.button
            { className: "btn btn-secondary py-1" <> guard props.disabled " disabled"
            , title: props.title
            , onClick: capture_ $ guard (not props.disabled) props.onClick
            , children: [ props.content ]
            }
        , R.button
            { className: "btn btn-secondary py-1 dropdown-toggle dropdown-toggle-split"
            , title: props.entriesTitle
            , _data: O.fromHomogeneous { toggle: "dropdown" }
            }
        , R.div
            { className: "dropdown-menu"
            , children:
              map
                ( \(Tuple content action) ->
                    R.a
                      { className: "dropdown-item"
                      , href: "#"
                      , onClick: capture_ $ action
                      , children: [ content ]
                      }
                )
                props.entries
            }
        ]
      }
  where
  component :: Component Props
  component = createComponent "DropDownButton"
