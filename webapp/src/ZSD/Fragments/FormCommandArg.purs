module ZSD.Fragments.FormCommandFlag where

import Data.Maybe (fromMaybe)
import Effect (Effect)
import Prelude (Unit)
import React.Basic (JSX)
import React.Basic.DOM as R
import React.Basic.DOM.Events (targetChecked)
import React.Basic.Events (handler)


flag :: String -> String -> (Boolean -> Effect Unit) -> JSX
flag name desc onChange =
  R.div
  { className: "form-check my-2"
  , children:
    [ R.input
      { className: "form-check-input"
      , type: "checkbox"
      , id: name
      , onChange: handler targetChecked \v -> onChange (fromMaybe false v)
      }
    , R.label
      { className: "form-check-label"
      , htmlFor: name
      , children:
        [ R.b_ [ R.text name ]
        , R.text " : "
        , R.text desc
        ]
      }
    ]
  }
