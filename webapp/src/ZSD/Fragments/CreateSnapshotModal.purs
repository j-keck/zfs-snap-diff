module ZSD.Fragments.CreateSnapshotModal where

import Prelude

import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), isJust)
import Data.Monoid (guard)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (Component, JSX, Self, createComponent, make)
import React.Basic.DOM as R
import ZSD.Components.Confirm as Confirm
import ZSD.Components.Spinner as Spinner
import ZSD.Fragments.SnapshotNameForm as SnapshotNameForm
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.Dataset as Dataset
import ZSD.Views.Messages as Messages

type Props
  = { dataset :: Dataset
    , snapshotNameTemplate :: String
    , onRequestClose :: Effect Unit
    }

type State
  = { snapshotName :: Maybe String }

data Action
  = CreateSnapshot

update :: Self Props State -> Action -> Effect Unit
update self = case _ of
  CreateSnapshot ->
    flip foldMap self.state.snapshotName \name ->
      Spinner.display
      *> launchAff_ do
          res <- Dataset.createSnapshot self.props.dataset name
          liftEffect $ do
            either Messages.appError Messages.info res
            self.props.onRequestClose
            Spinner.remove


createSnapshotModal :: Props -> JSX
createSnapshotModal = make component { initialState, render }
  where
  component :: Component Props
  component = createComponent "SnapshotNameModal"

  initialState = { snapshotName: Nothing }

  render self =
    Confirm.confirm
    { header: R.text "Create ZFS Snapshot"
    , body: SnapshotNameForm.snapshotNameForm
              { dataset: self.props.dataset
              , defaultTemplate: self.props.snapshotNameTemplate
              , onNameChange: \n -> self.setState _ { snapshotName = n }
              , onEnter: \_ -> update self CreateSnapshot
              , onEsc: self.props.onRequestClose
              }
    , onOk: guard (isJust self.state.snapshotName) $ update self CreateSnapshot
    , onCancel: self.props.onRequestClose
    }
