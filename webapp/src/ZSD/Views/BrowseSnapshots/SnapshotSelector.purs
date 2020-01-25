module ZSD.Views.BrowseSnapshots.SnapshotSelector where

import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Prelude (Unit, ($), discard)
import React.Basic (Component, JSX, createComponent, empty, make)
import React.Basic.DOM as R
import ZSD.Component.TableX (tableX)
import ZSD.Components.Panel (panel)
import ZSD.Formatter as Formatter
import ZSD.Model.Snapshot (Snapshots, Snapshot)

type Props =
  { snapshots :: Snapshots
  , onSnapshotSelected :: Snapshot -> Effect Unit
  }

type State =
  { selectedIdx :: Maybe Int}


snapshotSelector :: Props -> JSX
snapshotSelector = make component { initialState, render }
  where

    component :: Component Props
    component = createComponent "SelectSnapshot"

    initialState = { selectedIdx: Nothing }

    render self =
      panel
      { header: R.text "Snapshots"
      , body: \hidePanelBodyFn ->
        tableX
         { header: ["Snapshot Name", "Snapshot Created"]
         , rows: self.props.snapshots
         , mkRow: \s -> [ R.text s.name, R.text $ Formatter.dateTime s.created ]
         , onRowSelected: \(Tuple idx snapshot) -> do
             hidePanelBodyFn
             self.setState _ { selectedIdx = Just idx }
             self.props.onSnapshotSelected snapshot
         , activeIdx: self.state.selectedIdx
         }
      , footer: empty
      }

