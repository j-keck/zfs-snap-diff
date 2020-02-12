module ZSD.Components.Spinner
       ( display
       , remove
       , spinner
       ) where

import Prelude

import Data.Monoid (guard)
import Effect (Effect)
import Effect.Ref (Ref)
import Effect.Ref as Ref
import Effect.Timer (setInterval)
import Effect.Unsafe (unsafePerformEffect)
import Foreign.Object as O
import React.Basic (JSX, createComponent, fragment, make)
import React.Basic.DOM as R

display :: Effect Unit
display = Ref.write true visibleFlag

remove :: Effect Unit
remove = Ref.write false visibleFlag

spinner :: JSX
spinner = unit # make component { initialState, didMount, render }

  where
    component = createComponent "Spinner"

    initialState = { visible: false }

    didMount self = void $ setInterval 250 do
      flag <- Ref.read visibleFlag
      self.setState _ { visible = flag }

    render self =
      guard self.state.visible $ fragment
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




visibleFlag :: Ref Boolean
visibleFlag = unsafePerformEffect $ Ref.new false
