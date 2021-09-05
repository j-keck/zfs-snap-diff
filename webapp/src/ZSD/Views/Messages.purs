module ZSD.Views.Messages
  ( info
  , warning
  , error
  , appError
  , messages
  , toasts
  ) where

import Prelude
import Data.Array as A
import Data.DateTime (DateTime, adjust)
import Data.Time.Duration (Milliseconds(..), Seconds(..))
import Effect (Effect)
import Effect.Aff (delay, launchAff_)
import Effect.Class (liftEffect)
import Effect.Now (nowDateTime)
import Effect.Ref (Ref)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)
import React.Basic (JSX)
import React.Basic.Classic (createComponent, make, readState)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Components.Table (table)
import ZSD.Components.Spinner as Spinner
import ZSD.Model.AppError (AppError)
import ZSD.Utils.Ops (unsafeFromJust)

type Message
  = { ts :: DateTime, level :: Level, msg :: String }

data Level
  = Info
  | Warning
  | Error

derive instance eqLevel :: Eq Level

-- | Info message
info :: String -> Effect Unit
info = push Info

-- | Warning message
warning :: String -> Effect Unit
warning = push Warning

-- | Error message
error :: String -> Effect Unit
error = push Error

-- | AppError message
appError :: AppError -> Effect Unit
appError = push Error <<< show

push :: Level -> String -> Effect Unit
push level msg = do
  Spinner.remove
  ts <- nowDateTime
  Ref.modify_ (_ `A.snoc` { ts, level, msg }) history

history :: Ref (Array Message)
history = unsafePerformEffect $ Ref.new []

toasts :: JSX
toasts = unit # make component { initialState, didMount, render }
  where
  component = createComponent "Messages.toasts"

  initialState = { taken: 0, msgs: [] :: Array Message }

  didMount = update

  update self = do
    state <- readState self
    newMsgs <- A.drop state.taken <$> Ref.read history
    ts <- unsafeFromJust <<< adjust (Seconds $ negate 5.0) <$> nowDateTime
    self.setStateThen
      ( _
          { taken = state.taken + A.length newMsgs
          , msgs =
            A.concat
              [ A.filter (\m -> m.ts >= ts) state.msgs
              , newMsgs
              ]
          }
      )
      $ launchAff_
      $ delay (Milliseconds 500.0)
      *> liftEffect (update self)

  render self =
    R.div
      { className: "fixed-bottom px-5"
      , children: map (mkAlertBox self) self.state.msgs
      }

  mkAlertBox self n@{ ts, level, msg } =
    R.div
      { className: "alert alert-" <> level2bs level <> " fade show"
      , children:
        [ R.text msg
        , R.button
            { className: "close"
            , onClick: capture_ $ self.setState \s -> s { msgs = A.filter ((/=) n) s.msgs }
            , dangerouslySetInnerHTML: { __html: "&times;" }
            }
        ]
      }

messages :: JSX
messages = unit # make component { initialState, didMount, render }
  where
  component = createComponent "Messages.messages"

  initialState :: Array Message
  initialState = []

  didMount = update

  update self = do
    msgs <- Ref.read history
    self.setState (const msgs)
    launchAff_ $ delay (Milliseconds 500.0) *> liftEffect (update self)

  render self =
    table
      { header: [ R.text "Level", R.text "Message" ]
      , rows: self.state
      , mkRow:
        \{ ts, level, msg } ->
          [ R.span
              { className: "badge badge-" <> level2bs level
              , children: [ R.text $ level2name level ]
              }
          , R.text msg
          ]
      , onRowSelected: const $ pure unit
      }

-- | returns the level name
level2name :: Level -> String
level2name = case _ of
  Info -> "Info"
  Warning -> "Warning"
  Error -> "Error"

-- | returns the boostrap contextual varations name
level2bs :: Level -> String
level2bs = case _ of
  Info -> "secondary"
  Warning -> "warning"
  Error -> "danger"
