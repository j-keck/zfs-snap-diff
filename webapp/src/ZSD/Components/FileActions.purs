module ZSD.Components.FileActions where

import Prelude

import Data.Array as A
import Data.Either (either, fromRight)
import Data.Monoid (guard)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (logShow)
import Effect.Exception (throw)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, empty, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Components.LogLifecycles (logLifecycles)
import React.Basic.DOM.Events (capture_)
import Web.HTML (window)
import Web.HTML.Location (assign)
import Web.HTML.Window (location)
import ZSD.Components.FileAction.ViewDiff (viewDiff)
import ZSD.Components.FileActions.ViewBlob (viewBlob)
import ZSD.Components.FileActions.ViewText (viewText)
import ZSD.Model.Diff as Diff
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FSEntry as FSEntry
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.FileVersion as FileVersion
import ZSD.Model.MimeType (MimeType)
import ZSD.Model.MimeType as MimeType


type Props = { file :: FSEntry, version :: FileVersion, mimeType :: MimeType }
type State = { view :: JSX }

data Command =
    View
  | ViewText
  | ViewBlob
  | Diff
  | Download
  | Restore


update :: React.Self Props State -> Command -> Effect Unit
update self = case _ of

  View -> if(MimeType.isText self.props.mimeType)
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


  Diff -> launchAff_ $ do
    --diff <- unsafePartial $ fromRight <$> Diff.fetch self.props.file self.props.version
    --liftEffect $ self.setState _ { view = viewDiff { diff } }

    res <- Diff.fetch self.props.file self.props.version
    liftEffect $ do
      logShow res
      self.setState _ { view = viewDiff { diff: either (const mempty) identity res } }


  Download -> do
    let path = (FileVersion.unwrapFile self.props.version).path
        name = FileVersion.uniqueName self.props.version

    location <- window >>= location
    assign ("/api/download?path=" <> path <> "&name=" <> name) location

  Restore -> throw "restore function missing"


fileAction :: Props -> JSX
fileAction = logLifecycles <<< make component { initialState, render, didMount, didUpdate }

  where

    component :: Component Props
    component = createComponent "FileAction"

    initialState = { view: empty }

    didMount self = update self View

    didUpdate self { prevProps } = do
      guard (self.props /= prevProps) $ update self View


    render self =
      R.div
      { children:
        [ R.div
          { className: "btn-group"
          , children:
            [ btn "View" View $ A.any (\f -> f self.props.mimeType)
                                      [MimeType.isText, MimeType.isImage, MimeType.isPDF]
            , btn "Diff" Diff $ MimeType.isText self.props.mimeType
            , btn' "Download" Download
            -- FIXME: restore
            , btn "Restore" Restore false
            ]
          }
        , R.div
          { className: "card"
          , children:
            [ R.div
              { className: "card-header"
              , children: case self.props.version of
                             ActualVersion { name } -> [ R.text "Actual version from file: "
                                                      , R.text name
                                                      ]
                             BackupVersion { file, snapshot } -> [ R.text file.name
                                                                , R.text " from snapshot: "
                                                                , R.text snapshot.name
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

        btn title action enabled = R.button
                            { className: "btn btn-secondary" <> guard (not enabled) " disabled"
                            , onClick: capture_ $ if(enabled)
                                                  then update self action
                                                  else pure unit
                            , children: [ R.text title ]
                            }
        btn' title action = btn title action true
