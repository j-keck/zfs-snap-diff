module ZSD.Views.BrowseSnapshots.SnapshotSelector where

import Effect (Effect)
import Prelude (Unit, ($), discard)
import React.Basic (Component, JSX, createComponent, empty, make)
import React.Basic.DOM as R
import ZSD.Component.Table (table)
import ZSD.Components.Panel (panel)
import ZSD.Formatter as Formatter
import ZSD.Model.Snapshot (Snapshots, Snapshot)

type Props =
  { snapshots :: Snapshots
  , onSnapshotSelected :: Snapshot -> Effect Unit
  }

type State = {}


snapshotSelector :: Props -> JSX
snapshotSelector = make component { initialState, render }
  where

    component :: Component Props
    component = createComponent "SelectSnapshot"

    initialState = {}

    render self =
      panel
      { header: R.text "Snapshots"
      , body: \hidePanelBodyFn ->
        table
         { header: ["Snapshot Name", "Snapshot Created"]
         , rows: self.props.snapshots
         , mkRow: \s -> [ R.text s.name, R.text $ Formatter.dateTime s.created ]
         , onRowSelected: \snapshot -> do
             hidePanelBodyFn
             self.props.onSnapshotSelected snapshot
         }
      , footer: empty
      }

