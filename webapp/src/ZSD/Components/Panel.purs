module ZSD.Components.Panel where

import Prelude

import Data.Monoid (guard)
import Effect.Console (log, logShow)
import Effect.Ref (Ref)
import Effect.Ref as Ref
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Components.LogLifecycles (logLifecycles)
import React.Basic.DOM.Events (capture_)


type Props =
  { title :: String
  , body :: JSX
  , showBody :: Ref Boolean
  }

type State = { showBody :: Boolean }


panel :: Props -> JSX
panel = logLifecycles <<< make component { initialState, render, didUpdate }

  where

    component :: Component Props
    component = createComponent "Panel"

    initialState = { showBody: true }


    didUpdate self { prevProps, prevState } = do
      changed <- (/=) <$> Ref.read self.props.showBody <*> Ref.read prevProps.showBody
      guard changed
       (Ref.read prevProps.showBody >>= \b -> self.setState _ { showBody = b })

    render self =
      R.div
      { className: "card mt-3"
      , children:
        [ header self
        , guard self.state.showBody (R.div { className: "card-body", children: [self.props.body] })
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
