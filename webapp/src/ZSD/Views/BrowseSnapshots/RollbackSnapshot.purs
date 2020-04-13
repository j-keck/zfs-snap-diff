module ZSD.BrowseSnapshots.RollbackSnapshot where

import Prelude
import Data.Array as A
import Data.Either (either)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (Component, JSX, Self, createComponent, fragment, make)
import React.Basic.DOM as R
import ZSD.Components.Confirm as Confirm
import ZSD.Fragments.FormCommandFlag (flag)
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

type RollbackSnapshotArgs
  = Array String

type State
  = RollbackSnapshotArgs

data Action
  = RollbackSnapshot Snapshot

update :: Self Props State -> Action -> Effect Unit
update self = case _ of
  RollbackSnapshot snap ->
    launchAff_ do
      res <- Dataset.rollbackSnapshot self.props.dataset snap self.state
      liftEffect do
        either Messages.appError Messages.info res
        self.props.onOk

rollbackSnapshot :: Props -> JSX
rollbackSnapshot = make component { initialState, render }
  where
  component :: Component Props
  component = createComponent "RollbackSnapshot"

  initialState = []

  render self =
    Confirm.confirm
      { header: R.text "Rollback snapshot"
      , body:
        R.form
          { children:
            [ R.p { className: "font-weight-bold", children: [ R.text self.props.snap.fullName ] }
            , flag "-R"
                "Destroy any more recent snapshots and bookmarks, as well as any clones of those snapshots."
                (addOrRemoveFlag self "-R")
            , flag "-f"
                "Used with the -R option to force an unmount of any clone file systems that are to be destroyed."
                (addOrRemoveFlag self "-f")
            , flag "-r"
                "Destroy any snapshots and bookmarks more recent than the one specified."
                (addOrRemoveFlag self "-r")
            ]
          }
      , onOk: update self (RollbackSnapshot self.props.snap)
      , onCancel: self.props.onCancel
      }

  addOrRemoveFlag self flag true = self.setState (const $ A.snoc self.state flag)

  addOrRemoveFlag self flag false = self.setState (const $ A.filter ((/=) flag) self.state)
