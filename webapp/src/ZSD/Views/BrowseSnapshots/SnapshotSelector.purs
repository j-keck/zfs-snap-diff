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
import React.Basic (Component, JSX, createComponent, empty, fragment, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
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
    }

type State
  = { snapshots :: Snapshots
    , selectedIdx :: Maybe Int
    , modal :: JSX
    }

data Command
  = FetchSnapshots
  | SelectSnapshotByIdx Int
  | DestroySnapshot String

update :: React.Self Props State -> Command -> Effect Unit
update self = case _ of
  FetchSnapshots ->
    Spinner.display
      *> launchAff_
          ( Snapshots.fetchForDataset self.props.dataset
              >>= either Messages.appError (\snaps -> self.setState _ { snapshots = snaps } *> Spinner.remove)
              >>> liftEffect
          )
  SelectSnapshotByIdx idx ->
    Spinner.display
      *> self.setState _ { selectedIdx = Just idx }
      *> foldMap self.props.onSnapshotSelected (A.index self.state.snapshots idx)
      *> Spinner.remove
  DestroySnapshot name ->
    launchAff_ do
      res <- Dataset.destroySnapshot self.props.dataset name
      liftEffect do
        either Messages.appError Messages.info res
        update self FetchSnapshots

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
                { header: [ "Snapshot Name", "Snapshot Created" ]
                , rows: self.state.snapshots
                , mkRow:
                  \s ->
                    [ R.text s.name
                    , fragment
                        [ R.text $ Formatter.dateTime s.created
                        , R.span
                            { className: "float-right fas fa-trash pointer p-1"
                            , title: "Destroy snapshot"
                            , onClick:
                              capture_
                                $ self.setState
                                    _
                                      { modal =
                                        Confirm.confirm
                                          { header: R.text "Confirm destructive action"
                                          , body:
                                            R.text "Destroy snapshot "
                                              <> R.br {}
                                              <> R.b_ [ R.text s.fullName ]
                                          , onOk:
                                            self.setState _ { modal = empty }
                                              *> update self (DestroySnapshot s.name)
                                          , onCancel: self.setState _ { modal = empty }
                                          }
                                      }
                            }
                        ]
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

hasNewerSnapshots :: State -> Boolean
hasNewerSnapshots state = state.selectedIdx /= Just 0

hasOlderSnapshots :: State -> Boolean
hasOlderSnapshots state = fromMaybe 0 state.selectedIdx < A.length state.snapshots
