module ZSD.Components.DatasetSelector where

import Prelude
import Data.Maybe (Maybe(..))
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
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
datasetSelector = make component { initialState, render }

  where

    component :: Component Props
    component  = createComponent "DatasetSelector"

    initialState = { selectedDataset: Nothing }

    render self =
      panel
      { title: "Datasets"
      , body: \hidePanelBodyFn ->
          table
            { header: ["Name", "Used", "Avail", "Refer", "Mountpoint"]
            , rows: self.props.datasets
            , mkRow: \ds -> [ R.text ds.name
                            , R.text $ Formatter.filesize ds.used
                            , R.text $ Formatter.filesize ds.avail
                            , R.text $ Formatter.filesize ds.refer
                            , R.text ds.mountPoint.path ]
            , onRowSelected: \ds -> do
                    hidePanelBodyFn
                    self.setState _ { selectedDataset = Just ds }
                    self.props.onDatasetSelected ds
            }
      }
