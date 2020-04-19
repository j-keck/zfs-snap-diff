module ZSD.Fragments.FileActions where

import Prelude

import Data.Array as A
import Data.Either (either)
import Data.Monoid (guard)
import Data.Newtype (unwrap)
import Data.String as S
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (Component, JSX, createComponent, empty, make, readState)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import Web.HTML (window)
import Web.HTML.Location (assign)
import Web.HTML.Window (location)
import ZSD.Components.ActionButton (actionButton)
import ZSD.Fragments.FileAction.ViewDiff (viewDiff)
import ZSD.Fragments.FileActions.ViewBlob (viewBlob)
import ZSD.Fragments.FileActions.ViewText (viewText)
import ZSD.Model.FH (FH, fetchMimeType)
import ZSD.Model.FH as FH
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.FileVersion as FileVersion
import ZSD.Model.MimeType (MimeType(..))
import ZSD.Model.MimeType as MimeType
import ZSD.Utils.Ops (checkAny)
import ZSD.Views.Messages as Messages

type Props
  = { file :: FH, version :: FileVersion }

type State
  = { view :: JSX, cmd :: Command, mimeType :: MimeType }

data Command
  = View
  | ViewText
  | ViewBlob
  | Diff
  | Download
  | Restore

derive instance eqCommand :: Eq Command

update :: React.Self Props State -> Command -> Effect Unit
update self = case _ of
  View -> do
    state <- readState self
    self.setStateThen _ { cmd = View }
      $ if (MimeType.isText state.mimeType) then
          update self ViewText
        else
          update self ViewBlob
  ViewText ->
    launchAff_
      $ do
          res <- FH.downloadText (file self.props.version)
          liftEffect $ either Messages.appError (\content -> self.setState _ { view = viewText { content } }) res
  ViewBlob -> do
    mimeType <- _.mimeType <$> readState self
    if (checkAny [ MimeType.isPDF, MimeType.isImage ] mimeType) then
      launchAff_ do
        res <- FH.downloadBlob (file self.props.version)
        liftEffect $ either Messages.appError (\content -> self.setState _ { view = viewBlob { content } }) res
    else
      self.setState _ { view = R.text $ show mimeType <> " not embeddable" }
  Diff -> do
    self.setState
      _
        { view =
          viewDiff
            { file: self.props.file
            , version: self.props.version
            }
        , cmd = Diff
        }
  Download -> do
    let
      p = (unwrap >>> _.path) (file self.props.version)

      asName = FileVersion.uniqueName self.props.version
    location <- window >>= location
    assign ("api/download?path=" <> p <> "&as-name=" <> asName) location
  Restore ->
    launchAff_
      $ do
          res <- FileVersion.restore self.props.version
          liftEffect
            $ do
                either Messages.appError Messages.info res
                update self View

fileAction :: Props -> JSX
fileAction = make component { initialState, render, didMount, didUpdate }
  where
  component :: Component Props
  component = createComponent "FileAction"

  initialState = { view: empty, cmd: View, mimeType: MimeType "" }

  didMount self =
    launchAff_
      $ do
          res <- fetchMimeType self.props.file
          liftEffect $ either Messages.appError (\mt -> self.setStateThen _ { mimeType = mt } (update self View)) res

  didUpdate self { prevProps } = do
    guard (self.props /= prevProps) $ update self self.state.cmd

  render self =
    R.div
      { className: "mt-3"
      , children:
        [ R.div
            { className: "btn-group"
            , children:
              [ btn "View" "fas fa-eye" View
                  $ A.any (\f -> f self.state.mimeType) [ MimeType.isText, MimeType.isImage, MimeType.isPDF ]
              , btn "Diff" "fas fa-random" Diff $ MimeType.isText self.state.mimeType
              , btn "Download" "fas fa-download" Download true
              , actionButton
                  { text: "Restore"
                  , icon: "fas fa-archive"
                  , textConfirm: "Restore the old version of " <> (unwrap self.props.file).name
                  , action: update self Restore
                  , enabled: isBackupVersion self.props.version
                  }
              ]
            }
        , R.div
            { className: "card"
            , children:
              [ R.div
                  { className: "card-header"
                  , id: "file-actions-header"
                  , children:
                    case self.props.version of
                      CurrentVersion current ->
                        [ R.b_ [ R.text "Current" ]
                        , R.text " content of "
                        , R.b_ [ R.text $ (unwrap >>> _.name) current ]
                        ]
                      BackupVersion { backup, snapshot } ->
                        [ R.text "Content of "
                        , R.b_ [ R.text $ (unwrap >>> _.name) backup ]
                        , R.text " from snapshot: "
                        , R.b_ [ R.text snapshot.name ]
                        ]
                  }
              , R.div
                  { className: "card-body"
                  , id: "file-actions-body"
                  , children: [ self.state.view ]
                  }
              ]
            }
        ]
      }
    where
    btn title icon action enabled =
      R.button
        { className:
          "btn btn-secondary" <> guard (not enabled) " disabled"
            <> guard (self.state.cmd == action) " active"
        , id: "btn-" <> S.toLower title
        , onClick:
          capture_
            $ if (enabled) then
                update self action
              else
                pure unit
        , children:
          [ R.span { className: icon <> " p-1" }
          , R.text title
          ]
        }

    isBackupVersion = case _ of
      BackupVersion _ -> true
      CurrentVersion _ -> false

file :: FileVersion -> FH
file = case _ of
  CurrentVersion current -> current
  BackupVersion { backup } -> backup
