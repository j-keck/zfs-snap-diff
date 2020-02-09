module ZSD.Fragments.DatasetSelector where

import Prelude

import Data.Maybe (Maybe(..))
import Data.Newtype (unwrap)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, empty, make)
import React.Basic.DOM as R

import ZSD.Components.Panel (panel)
import ZSD.Components.Scroll as Scroll
import ZSD.Components.TableX (tableX)
import ZSD.Utils.Formatter as Formatter
import ZSD.Model.Dataset (Datasets, Dataset)


type Props =
  { datasets          :: Datasets
  , onDatasetSelected :: Dataset -> Effect Unit
  }

type State =
  { selectedDataset :: Maybe Dataset
  , activeIdx :: Maybe Int
  }

datasetSelector :: Props -> JSX
datasetSelector = make component { initialState, render }

  where

    component :: Component Props
    component  = createComponent "DatasetSelector"

    initialState = { selectedDataset: Nothing, activeIdx: Nothing }

    render self =
      panel
      { header: R.text "Datasets"
      , body: \hidePanelBodyFn ->
          tableX
            { header: ["Name", "Used", "Avail", "Refer", "Mountpoint"]
            , rows: self.props.datasets
            , mkRow: \ds -> [ R.text ds.name
                            , R.text $ Formatter.filesize ds.used
                            , R.text $ Formatter.filesize ds.avail
                            , R.text $ Formatter.filesize ds.refer
                            , R.text (unwrap ds.mountPoint).path ]
            , onRowSelected: \(Tuple idx ds) -> do
                    hidePanelBodyFn
                    Scroll.scrollToTop
                    self.setState _ { selectedDataset = Just ds, activeIdx = Just idx }
                    self.props.onDatasetSelected ds
            , activeIdx: self.state.activeIdx
            }
      , showBody: true
      , footer: empty
      }
