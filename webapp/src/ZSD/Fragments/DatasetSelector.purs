module ZSD.Fragments.DatasetSelector where

import Prelude
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), isNothing, maybe)
import Data.Newtype (unwrap)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, empty, fragment, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Components.Panel (panel)
import ZSD.Components.Scroll as Scroll
import ZSD.Components.TableX (tableX)
import ZSD.Fragments.CreateSnapshotModal (createSnapshotModal)
import ZSD.Model.Dataset (Datasets, Dataset)
import ZSD.Utils.Formatter as Formatter

type Props
  = { datasets :: Datasets
    , snapshotNameTemplate :: String
    , activeDataset :: Maybe Dataset
    , onDatasetSelected :: Dataset -> Effect Unit
    }

type State
  = { selectedDataset :: Maybe Dataset
    , activeIdx :: Maybe Int
    , modal :: JSX
    }

datasetSelector :: Props -> JSX
datasetSelector = make component { initialState, didMount, render }
  where
  component :: Component Props
  component = createComponent "DatasetSelector"

  initialState = { selectedDataset: Nothing, activeIdx: Nothing, modal: empty }

  didMount self = self.setState _ { selectedDataset = self.props.activeDataset }

  render self =
    panel
      { header:
        fragment
          [ R.text $ maybe "Datasets" (\{ name } -> "Dataset: " <> name) self.state.selectedDataset
          , foldMap
              ( \ds ->
                  R.span
                    { className: "float-right fas fa-camera pointer p-1"
                    , title: "Create a snapshot for " <> ds.name
                    , id: "create-snapshot"
                    , onClick:
                      capture_
                        $ self.setState
                            _
                              { modal =
                                createSnapshotModal
                                  { dataset: ds
                                  , snapshotNameTemplate: self.props.snapshotNameTemplate
                                  , onRequestClose:
                                    self.setState _ { modal = empty }
                                      *> self.props.onDatasetSelected ds
                                  }
                              }
                    }
              )
              self.state.selectedDataset
          ]
      , body:
        \hidePanelBodyFn ->
          tableX
            { header: [ "Name", "Used", "Avail", "Refer", "Mountpoint" ]
            , rows: self.props.datasets
            , mkRow:
              \ds ->
                [ R.text ds.name
                , R.text $ Formatter.filesize ds.used
                , R.text $ Formatter.filesize ds.avail
                , R.text $ Formatter.filesize ds.refer
                , R.text (unwrap ds.mountPoint).path
                ]
            , onRowSelected:
              \(Tuple idx ds) -> do
                hidePanelBodyFn
                Scroll.scrollToTop
                self.setState _ { selectedDataset = Just ds, activeIdx = Just idx }
                self.props.onDatasetSelected ds
            , activeIdx: self.state.activeIdx
            }
      , showBody: isNothing self.props.activeDataset
      , footer: empty
      }
      <> self.state.modal
