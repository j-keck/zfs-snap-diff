-- | A `ActionButton` executes the given action after a confirmed click
module ZSD.Components.ActionButton where

import Prelude
import Effect (Effect)
import Effect.Timer (setTimeout)
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, make, Self)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import Data.Monoid (guard)

type Props
  = { text :: String
    , title :: String
    , textConfirm :: String
    , icon :: String
    , action :: Effect Unit
    , enabled :: Boolean
    }

data State
  = Clean
  | Clicked
  | Confirmed

update :: Self Props State -> State -> Effect Unit
update self = case _ of
  Clean -> self.setState (const Clean)
  Clicked ->
    guard self.props.enabled do
      _ <- setTimeout 3000 $ update self Clean
      self.setState $ const Clicked
  Confirmed -> self.setStateThen (const Clean) $ self.props.action

actionButton :: Props -> JSX
actionButton = make component { initialState, render }
  where
  component :: Component Props
  component = createComponent "ActionButton"

  initialState = Clean

  render self =
    let
      className = "btn btn-secondary"
    in
      case self.state of
        Clean ->
          R.button
            { className: className <> guard (not self.props.enabled) " disabled"
            , title: self.props.title
            , onClick: capture_ $ update self Clicked
            , children:
              [ R.span { className: self.props.icon <> " p-1" }
              , R.text self.props.text
              ]
            }
        Clicked ->
          R.button
            { className: className <> " btn-warning"
            , title: self.props.title
            , onClick: capture_ $ update self Confirmed
            , children:
              [ R.span { className: "fas fa-exclamation p-1" }
              , R.text self.props.textConfirm
              ]
            }
        Confirmed -> mempty
