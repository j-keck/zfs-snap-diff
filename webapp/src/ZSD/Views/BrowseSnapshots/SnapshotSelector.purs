module ZSD.Views.BrowseSnapshots.SnapshotSelector where

import Prelude

import Data.Array as A
import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), fromMaybe, maybe)
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import Foreign.Object as O
import React.Basic (Component, JSX, createComponent, empty, fragment, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.BrowseSnapshots.CloneSnapshot as CloneSnapshot
import ZSD.BrowseSnapshots.DestroySnapshot as DestroySnapshot
import ZSD.BrowseSnapshots.RenameSnapshot as RenameSnapshot
import ZSD.BrowseSnapshots.RollbackSnapshot as RollbackSnapshot
import ZSD.Components.Confirm as Confirm
import ZSD.Components.Panel (panel)
import ZSD.Components.Spinner as Spinner
import ZSD.Components.TableX (tableX)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.Dataset as Dataset
import ZSD.Model.Snapshot (Snapshots, Snapshot)
import ZSD.Model.Snapshot as Snapshots
import ZSD.Utils.Formatter as Formatter
import ZSD.Views.Messages as Messages

type Props
  = { dataset :: Dataset
    , onSnapshotSelected :: Snapshot -> Effect Unit
    , onDatasetChanges :: Effect Unit
    }

type State
  = { snapshots :: Snapshots
    , selectedIdx :: Maybe Int
    , modal :: JSX
    }

data Command
  = FetchSnapshots
  | SelectSnapshotByIdx Int
  | DestroySnapshot Snapshot
  | RenameSnapshot Snapshot
  | CloneSnapshot Snapshot
  | RollbackSnapshot Snapshot

update :: React.Self Props State -> Command -> Effect Unit
update self = case _ of

  FetchSnapshots ->
    self.props.onDatasetChanges *>
    Spinner.display
      *> launchAff_
          ( Snapshots.fetchForDataset self.props.dataset.name
              >>= either Messages.appError (\snaps -> self.setState _ { snapshots = snaps } *> Spinner.remove)
              >>> liftEffect
          )

  SelectSnapshotByIdx idx ->
    Spinner.display
      *> self.setState _ { selectedIdx = Just idx }
      *> foldMap self.props.onSnapshotSelected (A.index self.state.snapshots idx)
      *> Spinner.remove


  DestroySnapshot snap ->
    self.setState _ {
      modal = DestroySnapshot.destroySnapshot
              { dataset: self.props.dataset
              , snap
              , onOk: update self FetchSnapshots *> self.setState _ { modal = empty }
              , onCancel: self.setState _ { modal = empty }
              }
      }

  RenameSnapshot snap ->
    self.setState _ {
      modal = RenameSnapshot.renameSnapshot
              { dataset: self.props.dataset
              , snap
              , onOk: update self FetchSnapshots *> self.setState _ { modal = empty }
              , onCancel: self.setState _ { modal = empty }
              }
      }

  CloneSnapshot snap ->
    self.setState _ {
      modal = CloneSnapshot.cloneSnapshot
              { dataset: self.props.dataset
              , snap
              , onOk: self.props.onDatasetChanges *> self.setState _ { modal = empty }
              , onCancel: self.setState _ { modal = empty }
              }
      }

  RollbackSnapshot snap ->
    self.setState _ {
      modal = RollbackSnapshot.rollbackSnapshot
              { dataset: self.props.dataset
              , snap
              , onOk: update self FetchSnapshots *> self.setState _ { modal = empty }
              , onCancel: self.setState _ { modal = empty }
              }
      }


snapshotSelector :: Props -> JSX
snapshotSelector = make component { initialState, didMount, didUpdate, render }
  where
  component :: Component Props
  component = createComponent "SelectSnapshot"

  initialState = { snapshots: [], selectedIdx: Nothing, modal: empty }

  didMount self = update self FetchSnapshots

  didUpdate self { prevProps, prevState } = guard (prevProps.dataset /= self.props.dataset) $ update self FetchSnapshots

  render self =
    fragment
      [ panel
          { header:
            fragment
              [ R.text "Snapshots"
              , R.span
                  { className: "float-right"
                  , children:
                    [ R.div
                        { className: "btn-group"
                        , children:
                          [ R.button
                              { className: "btn btn-secondary" <> guard (not $ hasOlderSnapshots self.state) " disabled"
                              , title: "Select the previous snapshot"
                              , onClick:
                                capture_ $ guard (hasOlderSnapshots self.state)
                                  $ update self (SelectSnapshotByIdx (maybe 0 (_ + 1) self.state.selectedIdx))
                              , children:
                                [ R.span { className: "fas fa-backward p-1" }
                                , R.text "Older"
                                ]
                              }
                          , R.button
                              { className: "btn btn-secondary" <> guard (not $ hasNewerSnapshots self.state) " disabled"
                              , title: "Select the successor snapshot"
                              , onClick:
                                capture_ $ guard (hasNewerSnapshots self.state)
                                  $ update self (SelectSnapshotByIdx (maybe 0 (_ - 1) self.state.selectedIdx))
                              , children:
                                [ R.text "Newer"
                                , R.span { className: "fas fa-forward p-1" }
                                ]
                              }
                          ]
                        }
                    ]
                  }
              ]
          , body:
            \hidePanelBodyFn ->
              tableX
                { header: [ "Snapshot Name", "Snapshot Created", "" ]
                , rows: self.state.snapshots
                , mkRow:
                  \s ->
                    [ R.text s.name
                    , R.text $ Formatter.dateTime s.created
                    , snapshotActions self s
                    ]
                , onRowSelected:
                  \(Tuple idx snapshot) -> do
                    hidePanelBodyFn
                    self.setState _ { selectedIdx = Just idx }
                    self.props.onSnapshotSelected snapshot
                , activeIdx: self.state.selectedIdx
                }
          , showBody: true
          , footer: empty
          }
      , self.state.modal
      ]

  snapshotActions self snap =
    R.div
    { className: "dropleft"
    , children:
      [ R.div
        { className: "dropdown"
        , _data: O.fromHomogeneous { toggle: "dropdown" }
        , onClick: capture_ $ pure unit -- prevent the row action
        , children:
          [ R.div
            { className: "mx-auto"
            , style: R.css { width: "10px" }
            , children: [ R.span { className: "fas fa-ellipsis-v" } ]
            }
          ]
        }
      , R.div
        { className: "dropdown-menu"
        , children:
          [ dropdownItem self "Rename snapshot" $ RenameSnapshot snap
          , dropdownItem self "Destroy snapshot" $ DestroySnapshot snap
          , dropdownItem self "Clone" $ CloneSnapshot snap
          , dropdownItem self "Rollback" $ RollbackSnapshot snap
          ]
        }
      ]
    }


  dropdownItem self name cmd =
    R.button
    { className: "dropdown-item"
    , onClick: capture_ $ update self cmd
    , children: [ R.text name ]
    }


hasNewerSnapshots :: State -> Boolean
hasNewerSnapshots state = state.selectedIdx /= Just 0

hasOlderSnapshots :: State -> Boolean
hasOlderSnapshots state = fromMaybe 0 state.selectedIdx < A.length state.snapshots
