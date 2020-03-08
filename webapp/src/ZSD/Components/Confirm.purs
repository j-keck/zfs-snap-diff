module ZSD.Components.Confirm where

import Prelude
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, fragment, makeStateless)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

type Props
  = { header :: JSX
    , body :: JSX
    , onOk :: Effect Unit
    , onCancel :: Effect Unit
    }

confirm :: Props -> JSX
confirm =
  makeStateless component \props ->
    fragment
    [ div "modal modal-show"
        $ div "modal-dialog modal-lg modal-dialog-centered"
        $ div "modal-content"
        $ fragment
            [ div "modal-header" $ props.header
            , div "modal-body" $ props.body
            , div "modal-footer"
                $ fragment
                    [ R.button
                        { className: "btn btn-secondary"
                        , onClick: capture_ props.onCancel
                        , children: [ R.text "Cancel" ]
                        }
                    , R.button
                        { className: "btn btn-primary"
                        , onClick: capture_ props.onOk
                        , children: [ R.text "Ok" ]
                        }
                    ]
            ]
    , R.div { className: "modal-backdrop fade show" }
    ]
  where
  component :: Component Props
  component = createComponent "Confirm"

  div className child = R.div { className, children: [ child ] }
