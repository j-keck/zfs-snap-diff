-- | A `ActionButton` executes the given action after a confirmed click
module ZSD.Components.ActionButton where

import Prelude

import Effect (Effect)
import Effect.Timer (setTimeout)
import React.Basic (Component, JSX, Self, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

type Props =
  { text :: String
  , textConfirm :: String
  , action :: Effect Unit
  }


data State =
    Clean
  | Clicked
  | Confirmed

update :: Self Props State -> State -> Effect Unit
update self = case _ of
  Clean -> self.setState (const Clean)
  Clicked -> do
    _ <- setTimeout 3000 $ update self Clean
    self.setState $ const Clicked
  Confirmed -> self.setStateThen (const Clean) $ self.props.action


actionButton :: Props -> JSX
actionButton = make component { initialState, render }
  where

    component :: Component Props
    component  = createComponent "ActionButton"

    initialState = Clean

    render self =
      let className = "btn btn-secondary" in
      case self.state of
        Clean -> R.button { className
                          , onClick: capture_ $ update self Clicked
                          , children: [ R.text self.props.text ]
                          }
        Clicked -> R.button { className: className <> " btn-warning"
                            , onClick: capture_ $ update self Confirmed
                            , children: [ R.text self.props.textConfirm ]
                            }
        Confirmed -> mempty

