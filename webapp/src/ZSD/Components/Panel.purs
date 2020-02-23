-- | A `Panel` component contains a panel-header and a panel-body.
-- | Header and body parts are provided in the `Props` record.
-- | The body can be hidden with the provided `HideBodyFn` function.
module ZSD.Components.Panel where

import Prelude
import Data.Monoid (guard)
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

-- | Type of the function to hide the body
type HideBodyFn
  = Effect Unit

-- | - `header`
-- |   - Panel header
-- | - `body`
-- |   - Panel body
type Props
  = { header :: JSX
    , body :: HideBodyFn -> JSX
    , showBody :: Boolean
    , footer :: JSX
    }

type State
  = { showBody :: Boolean }

panel :: Props -> JSX
panel = make component { initialState, didMount, render }
  where
  component :: Component Props
  component = createComponent "Panel"

  initialState = { showBody: true }

  didMount self = self.setState _ { showBody = self.props.showBody }

  render self =
    let
      hideBodyFn = self.setState _ { showBody = false }
    in
      R.div
        { className: "card mt-3"
        , children:
          [ R.h5
              { className: "card-header p-1"
              , style: R.css { cursor: "pointer" }
              , onClick: capture_ $ self.setState \s -> s { showBody = not s.showBody }
              , children:
                [ R.span
                    { className:
                      "p-1 fas fa-chevron-"
                        <> if self.state.showBody then
                            "up"
                          else
                            "down"
                    }
                , self.props.header
                ]
              }
          , R.div
              { className: "card-body pb-0" <> guard (not self.state.showBody) " d-none"
              , children: [ self.props.body hideBodyFn ]
              }
          , self.props.footer
          ]
        }
