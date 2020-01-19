module ZSD.Components.FileActions where

import Prelude

import Data.Array as A
import Data.Either (fromRight)
import Data.Monoid (guard)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, empty, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import Web.HTML (window)
import Web.HTML.Location (assign)
import Web.HTML.Window (location)
import ZSD.Components.ActionButton (actionButton)
import ZSD.Components.FileAction.ViewDiff (viewDiff)
import ZSD.Components.FileActions.ViewBlob (viewBlob)
import ZSD.Components.FileActions.ViewText (viewText)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FSEntry as FSEntry
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.FileVersion as FileVersion
import ZSD.Model.MimeType (MimeType(..))
import ZSD.Model.MimeType as MimeType


type Props = { file :: FSEntry, version :: FileVersion }
type State = { view :: JSX, mimeType :: MimeType }

data Command =
    View
  | ViewText
  | ViewBlob
  | Diff
  | Download
  | Restore


update :: React.Self Props State -> Command -> Effect Unit
update self = case _ of

  View -> if(MimeType.isText self.state.mimeType)
           then update self ViewText
           else update self ViewBlob


  ViewText -> launchAff_ $ do
    let file = FileVersion.unwrapFile self.props.version
    content <- unsafePartial $ fromRight <$> FSEntry.downloadText file
    liftEffect $ self.setState _ { view = viewText { content } }


  ViewBlob -> launchAff_ $ do
    let file = FileVersion.unwrapFile self.props.version
    content <- unsafePartial $ fromRight <$> FSEntry.downloadBlob file
    liftEffect $ self.setState _ { view = viewBlob { content } }


  Diff -> do
    let file = self.props.file
        version = self.props.version
    self.setState _ { view = viewDiff { file, version } }


  Download -> do
    let path = (FileVersion.unwrapFile self.props.version).path
        asName = FileVersion.uniqueName self.props.version

    location <- window >>= location
    assign ("/api/download?path=" <> path <> "&as-name=" <> asName) location

  Restore -> launchAff_ $ do
    _ <- FileVersion.restore self.props.file self.props.version
    liftEffect $ update self View


fileAction :: Props -> JSX
fileAction = make component { initialState, render, didMount, didUpdate }

  where

    component :: Component Props
    component = createComponent "FileAction"

    initialState = { view: empty, mimeType: MimeType "text/plain" }

    didMount self = launchAff_ $ do
      -- FIXME: handle error
      mimeType <- unsafePartial $ fromRight <$> MimeType.fetch self.props.file
      liftEffect $ self.setStateThen _ { mimeType = mimeType } (update self View)

    didUpdate self { prevProps } = do
      guard (self.props /= prevProps) $ update self View


    render self =
      R.div
      { className: "mt-3"
      , children:
        [ R.div
          { className: "btn-group"
          , children:
            [ btn "View" "fas fa-eye" View $
                A.any (\f -> f self.state.mimeType) [MimeType.isText, MimeType.isImage, MimeType.isPDF]

            , btn "Diff" "fas fa-random" Diff $
                (MimeType.isText self.state.mimeType) && (FileVersion.isBackupVersion self.props.version)

            , btn "Download" "fas fa-download" Download true
            , actionButton { text: "Restore"
                           , icon: "fas fa-archive"
                           , textConfirm: "Restore the old version of " <> self.props.file.name
                           , action: update self Restore
                           }
            ]
          }
        , R.div
          { className: "card"
          , children:
            [ R.div
              { className: "card-header"
              , children: case self.props.version of
                             ActualVersion { name } -> [ R.text "Actual content from: "
                                                      , R.b_ [ R.text name ]
                                                      ]
                             BackupVersion { file, snapshot } -> [ R.text "Content from: "
                                                                 , R.b_ [ R.text file.name ]
                                                                 , R.text " from snapshot: "
                                                                 , R.b_ [ R.text snapshot.name ]
                                                                ]
              }
            , R.div
              { className: "card-body"
              , children: [ self.state.view ]
              }
            ]
          }
        ]
      }

      where

        btn title icon action enabled =
          R.button
          { className: "btn btn-secondary" <> guard (not enabled) " disabled"
          , onClick: capture_ $ if(enabled)
                                then update self action
                                else pure unit
          , children:
            [ R.span { className: icon <> " p-1" }
            , R.text title
            ]
          }
