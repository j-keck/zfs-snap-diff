module ZSD.Components.DatasetSelector where

import Prelude
import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Components.LogLifecycles (logLifecycles)
import ZSD.Component.Table (table)
import ZSD.Components.Panel (panel)
import ZSD.Formatter as Formatter
import ZSD.Model.Dataset (Datasets, Dataset)


type Props =
  { datasets          :: Datasets
  , onDatasetSelected :: Dataset -> Effect Unit
  }

type State = { selectedDataset :: Maybe Dataset }

datasetSelector :: Props -> JSX
datasetSelector = logLifecycles <<< make component { initialState, render }

  where

    component :: Component Props
    component  = createComponent "Datasets"

    initialState = { selectedDataset: Nothing }

    render self =
      let showPanelBody = unsafePerformEffect $ Ref.new true in
      panel
      { title: "Datasets"
      , body: table
        { header: ["Name", "Used", "Avail", "Refer", "Mountpoint"]
        , rows: self.props.datasets
        , mkRow: \ds -> [ R.text ds.name
                       , R.text $ Formatter.filesize ds.used
                       , R.text $ Formatter.filesize ds.avail
                       , R.text $ Formatter.filesize ds.refer
                       , R.text ds.mountPoint.path ]
        , onRowSelected: \ds -> do
                Ref.write false showPanelBody
                self.setState _ { selectedDataset = Just ds }
                self.props.onDatasetSelected ds

        }
      , showBody: showPanelBody
      }
