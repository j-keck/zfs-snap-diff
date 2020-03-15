module ZSD.Fragments.CreateSnapshotModal where

import Prelude

import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), isNothing)
import Data.Monoid (guard)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (Component, JSX, Self, createComponent, fragment, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Fragments.SnapshotNameForm as SnapshotNameForm
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.Dataset as Dataset
import ZSD.Views.Messages as Messages

type Props
  = { dataset :: Dataset
    , snapshotNameTemplate :: String
    , onRequestClose :: Effect Unit
    }

type State = { snapshotName :: Maybe String }
data Action = CreateSnapshot


update :: Self Props State -> Action -> Effect Unit
update self = case _ of

  CreateSnapshot -> flip foldMap self.state.snapshotName \name ->  launchAff_ $
           Dataset.createSnapshot self.props.dataset name
      >>= (\res -> liftEffect $ either Messages.appError Messages.info res *> self.props.onRequestClose)

createSnapshotModal :: Props -> JSX
createSnapshotModal = make component { initialState, render }
  where
  component :: Component Props
  component = createComponent "SnapshotNameModal"

  initialState = { snapshotName: Nothing }

  render self =
    fragment
      [ R.div
          { className: "modal modal-show"
          , children:
            [ div "modal-dialog modal-dialog-centered"
                $ div "modal-content"
                $ fragment
                    [ div "modal-header" $ R.text "Create ZFS Snapshot"
                    , div "modal-body m-1" $ SnapshotNameForm.snapshotNameForm
                                             { dataset: self.props.dataset
                                             , defaultTemplate: self.props.snapshotNameTemplate
                                             , onNameChange: \n -> self.setState _ { snapshotName = n }
                                             }
                    , div "modal-footer"
                        $ fragment
                            [ R.button
                                { className: "btn btn-secondary"
                                , onClick: capture_ self.props.onRequestClose
                                , children: [ R.text "Cancel" ]
                                }
                            , R.button
                                { className:
                                  "btn btn-primary"
                                    <> guard (isNothing self.state.snapshotName) " disabled"
                                , onClick: capture_ $ update self CreateSnapshot
                                , children: [ R.text "Create" ]
                                }
                            ]
                    ]
            ]
          }
      , R.div { className: "modal-backdrop fade show" }
      ]

  div className child = R.div { className, children: [ child ] }
