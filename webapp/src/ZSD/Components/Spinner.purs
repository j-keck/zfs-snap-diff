module ZSD.Components.Spinner where

import Prelude
import React.Basic (makeStateless, createComponent, JSX, fragment)
import React.Basic.DOM as R
import Foreign.Object as O

spinner :: JSX
spinner = unit # makeStateless component \props ->
  fragment
  [ R.div
    { className: "modal show"
    , style: R.css { display: "block" }
    , _data: O.fromHomogeneous {backdrop: "static" }
    , tabIndex: -1
    , children:
      [ R.div
        { className: "modal-dialog modal-dialog-centered"
        , children:
          [ R.div
            { className: "modal-body text-center"
            , children:
              [ R.div
                { className: "spinner-border"
                , style: R.css { width: "6rem", height: "6rem" }
                , role: "status"
                , children:
                  [ R.span
                    { className: "sr-only"
                    , children: [ R.text "Loading ..." ]
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  , R.div
    { className: "modal-backdrop fade show" }
  ]
      

  where
    component = createComponent "Spinner"
