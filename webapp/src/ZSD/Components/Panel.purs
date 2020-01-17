-- | A `Panel` component contains a panel-header and a panel-body.
-- | Title and body parts are provided in the `Props` record.
-- | The body can be hidden with the provided `HideBodyFn` function.
module ZSD.Components.Panel where

import Prelude
import Data.Monoid (guard)
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

-- | Type of the function to hide the body
type HideBodyFn = Effect Unit

-- | - `title`
-- |   - Panel header title
-- | - `body`
-- |   - Panel body
type Props =
  { title :: String
  , body :: HideBodyFn -> JSX
  }

type State = { showBody :: Boolean }


panel :: Props -> JSX
panel = make component { initialState, render }

  where

    component :: Component Props
    component = createComponent "Panel"

    initialState = { showBody: true }

    render self = 
      let hideBodyFn = self.setState _ { showBody = false } in
      R.div
      { className: "card mt-3"
      , children:
        [ header self
        , guard self.state.showBody (R.div { className: "card-body", children: [self.props.body hideBodyFn ] })
        ]
      }

    header self =
      R.h5
      { className: "card-header p-1"
      , style: R.css { cursor: "pointer" }
      , onClick: capture_ $ self.setState \s -> s { showBody = not s.showBody }
      , children:
        [ R.img
          { className: "p-1"
          , src: if self.state.showBody
                 then "icons/chevron-up.svg"
                 else "icons/chevron-down.svg"
          , height: "32"
          }
        , R.text self.props.title
        ]
      }
