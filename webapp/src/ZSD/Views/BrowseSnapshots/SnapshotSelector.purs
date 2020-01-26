module ZSD.Views.BrowseSnapshots.SnapshotSelector where

import Data.Either (either)
import Data.Maybe (Maybe(..))
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Prelude (Unit, bind, const, discard, identity, ($), (/=))
import React.Basic (Component, JSX, createComponent, empty, fragment, make)
import React.Basic as React
import React.Basic.DOM as R
import ZSD.Components.Panel (panel)
import ZSD.Components.Spinner as Spinner
import ZSD.Components.TableX (tableX)
import ZSD.Formatter as Formatter
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.Snapshot (Snapshots, Snapshot)
import ZSD.Model.Snapshot as Snapshots

type Props =
  { dataset :: Dataset, onSnapshotSelected :: Snapshot -> Effect Unit }

type State =
  { snapshots :: Snapshots, selectedIdx :: Maybe Int, spinner :: JSX}


data Command =
  FetchSnapshots

update :: React.Self Props State -> Command -> Effect Unit
update self = case _ of
  FetchSnapshots ->
    self.setStateThen _ { spinner = Spinner.spinner } $ launchAff_ $ do
      res <- Snapshots.fetchForDataset self.props.dataset 
      liftEffect $ self.setState _ { snapshots = either (const []) identity res, spinner = empty }

snapshotSelector :: Props -> JSX
snapshotSelector = make component { initialState, didMount, didUpdate, render }
  where

    component :: Component Props
    component = createComponent "SelectSnapshot"

    initialState = { snapshots: [], selectedIdx: Nothing, spinner: empty }

    didMount self = update self FetchSnapshots

    didUpdate self { prevProps, prevState } = 
      guard (prevProps.dataset /= self.props.dataset) $ update self FetchSnapshots
        

    render self = fragment
      [ panel
        { header: R.text "Snapshots"
        , body: \hidePanelBodyFn ->
          tableX
           { header: ["Snapshot Name", "Snapshot Created"]
           , rows: self.state.snapshots
           , mkRow: \s -> [ R.text s.name, R.text $ Formatter.dateTime s.created ]
           , onRowSelected: \(Tuple idx snapshot) -> do
               hidePanelBodyFn
               self.setState _ { selectedIdx = Just idx }
               self.props.onSnapshotSelected snapshot
           , activeIdx: self.state.selectedIdx
           }
        , footer: empty
        }
      , self.state.spinner
      ]

