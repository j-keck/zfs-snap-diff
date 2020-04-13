module ZSD.BrowseSnapshots.RenameSnapshot where

import Prelude

import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (Component, JSX, Self, createComponent, fragment, make)
import React.Basic.DOM as R
import ZSD.Components.Confirm as Confirm
import ZSD.Fragments.SnapshotNameForm as SnapshotNameForm
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.Dataset as Dataset
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Views.Messages as Messages

type Props
  = { dataset :: Dataset
    , snap :: Snapshot
    , onOk :: Effect Unit
    , onCancel :: Effect Unit
    }

type State
  = { newName :: Maybe String }

data Action
  = RenameSnapshot Snapshot String

update :: Self Props State -> Action -> Effect Unit
update self = case _ of
  RenameSnapshot snap newName ->
    launchAff_ do
      res <- Dataset.renameSnapshot self.props.dataset snap newName
      liftEffect do
        either Messages.appError Messages.info res
        self.props.onOk

renameSnapshot :: Props -> JSX
renameSnapshot = make component { initialState, render }
  where
  component :: Component Props
  component = createComponent "RenameSnapshot"

  initialState = { newName: Nothing }

  render self =
    Confirm.confirm
      { header: R.text "Rename snapshot"
      , body:
        fragment
          [ R.b_ [ R.text self.props.snap.fullName ]
          , R.br {}
          , SnapshotNameForm.snapshotNameForm
              { dataset: self.props.dataset
              , defaultTemplate: self.props.snap.name
              , onNameChange: \name -> self.setState _ { newName = name }
              , onEnter: \name -> update self (RenameSnapshot self.props.snap name)
              , onEsc: self.props.onCancel
              }
          ]
      , onOk: flip foldMap self.state.newName \name -> update self (RenameSnapshot self.props.snap name)
      , onCancel: self.props.onCancel
      }
