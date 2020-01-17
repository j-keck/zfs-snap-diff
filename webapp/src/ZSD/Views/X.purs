module ZSD.Views.X where

import Prelude

import Effect.Console (log)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Model.FSEntry (FSEntry)


type Props = { file :: FSEntry }
type State = Int

x :: Props -> JSX 
x = make component { initialState, render, didMount }
  where

    component :: Component Props
    component = createComponent "X"

    initialState = 0

    didMount self = do
      log "X: didMount"
      self.setState (const 0)

    render self =
      R.div
      { onClick: capture_ $ self.setState ((+) 1)
      , children:
        [ R.text "Click count: "
        , R.text $ show self.state
        ]
      }


