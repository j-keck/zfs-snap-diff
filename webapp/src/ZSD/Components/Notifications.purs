module ZSD.Components.Notifications where

import Prelude

import Data.Array as A
import Data.Maybe (Maybe(..))
import Data.Time.Duration (Milliseconds(..))
import Effect (Effect)
import Effect.Aff (delay, launchAff_)
import Effect.Class (liftEffect)
import Effect.Ref (Ref)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)
import React.Basic (JSX, createComponent, fragment, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Model.AppError (AppError)


data Level =
    Info
  | Warning
  | Error
derive instance eqLevel :: Eq Level

type Notification =
  { level :: Level, msg :: String }

queue :: Ref (Array Notification)
queue = unsafePerformEffect $ Ref.new []


enqueue :: Level -> String -> Effect Unit
enqueue level msg =
  Ref.modify_ (_ `A.snoc` { level, msg }) queue

enqueueInfo :: String -> Effect Unit
enqueueInfo = enqueue Info
enqueueWarning :: String -> Effect Unit
enqueueWarning = enqueue Warning
enqueueError :: String -> Effect Unit
enqueueError = enqueue Error
enqueueAppError :: AppError -> Effect Unit
enqueueAppError = enqueue Error <<< show


notifications :: JSX
notifications = unit # make component { initialState, didMount, render }
  where
    component = createComponent "Notifications"

    initialState = {notifications: []}

    didMount = displayNotifications 

    displayNotifications self = launchAff_ do
      n <- liftEffect $ do
           q <- Ref.read queue
           Ref.write (A.drop 1 q) queue
           pure $ A.head q
      case n of
        Just n'-> liftEffect $ self.setStateThen (\s -> s { notifications = A.snoc s.notifications n' })
                                                 (displayNotifications self)
        Nothing -> delay (Milliseconds 500.0) *> liftEffect (displayNotifications self)


    render self =
      fragment $ map (mkAlertBox self) self.state.notifications


    mkAlertBox self n@{ level, msg } =
      R.div
      { className: "alert " <> level2class level <> " fade show"
      , children:
        [ R.text msg
        , R.button
          { className: "close"
          , onClick: capture_ $ self.setState \s -> s { notifications = A.filter ((/=) n) s.notifications }
          , dangerouslySetInnerHTML: { __html: "&times;" }
          }
        ]
      }


    level2class = case _ of
      Info -> "alert-secondary"
      Warning -> "alert-warning"
      Error -> "alert-danger"
