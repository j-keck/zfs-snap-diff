module ZSD.BrowseSnapshots.DestroySnapshot where

import Prelude

import Data.Array as A
import Data.Either (either)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (JSX)
import React.Basic.Classic (Component, Self, createComponent, make)
import React.Basic.DOM as R
import ZSD.Components.Confirm as Confirm
import ZSD.Components.Spinner as Spinner
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

type DestroySnapshotArgs
  = Array String

type State
  = DestroySnapshotArgs

data Action
  = DestroySnapshot Snapshot

update :: Self Props State -> Action -> Effect Unit
update self = case _ of
  DestroySnapshot snap ->
    Spinner.display *> launchAff_ do
      res <- Dataset.destroySnapshot self.props.dataset snap self.state
      liftEffect $ either Messages.appError Messages.info res *> self.props.onOk *> Spinner.remove

destroySnapshot :: Props -> JSX
destroySnapshot = make component { initialState, render }
  where
  component :: Component Props
  component = createComponent "DestroySnapshot"

  initialState = []

  render self =
    Confirm.confirm
      { header: R.text "Destroy snapshot"
      , body:
        R.form
          { children:
            [ R.p { className: "font-weight-bold", children: [ R.text self.props.snap.fullName ] }
            , flag "-R"
                "Recursively destroy all clones of these snapshots, including the clones, snapshots, and children.  If this flag is specified, the -d flag will have no effect."
                (addOrRemoveFlag self "-R")
            , flag "-d"
                "Destroy immediately. If a snapshot cannot be destroyed now, mark it for deferred destruction."
                (addOrRemoveFlag self "-d")
            , flag "-r"
                "Destroy (or mark for deferred deletion) all snapshots with this name in descendent file systems."
                (addOrRemoveFlag self "-r")
            ]
          }
      , onOk: update self (DestroySnapshot self.props.snap)
      , onCancel: self.props.onCancel
      }

  addOrRemoveFlag self flag true = self.setState (const $ A.snoc self.state flag)

  addOrRemoveFlag self flag false = self.setState (const $ A.filter ((/=) flag) self.state)
