module ZSD.Views.BrowseFilesystem where

import Prelude

import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.Tuple.Nested (tuple2, uncurry2)
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import ZSD.Components.DatasetSelector (datasetSelector)
import ZSD.Components.DirBrowser (dirBrowser)
import ZSD.Components.FileActions (fileAction)
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion)
import ZSD.Views.BrowseFilesystem.FileVersionSelector (fileVersionSelector)

type Props = { config :: Config }

type State = { selectedDataset      :: Maybe Dataset
             , selectedFile         :: Maybe FSEntry
             , selectedVersion      :: Maybe FileVersion
             }


data Command =
    DatasetSelected Dataset
  | FileSelected FSEntry
  | DirSelected FSEntry
  | VersionSelected FileVersion


update :: (React.Self Props State) -> Command -> Effect Unit
update self = case _ of
  DatasetSelected ds ->
    self.setState _ { selectedDataset = Just ds
                    , selectedFile = Nothing
                    , selectedVersion = Nothing
                    }

  FileSelected f ->
    self.setState _ { selectedFile = Just f
                    , selectedVersion = Nothing
                    }

  DirSelected d ->
    self.setState _ { selectedFile = Nothing
                    , selectedVersion = Nothing
                    }

  VersionSelected v -> self.setState _ { selectedVersion = Just v }



browseFilesystem :: Props -> JSX
browseFilesystem = make component { initialState, render }

  where

    component :: Component Props
    component = createComponent "BrowseFilesystem"

    initialState = { selectedDataset: Nothing
                   , selectedFile: Nothing
                   , selectedVersion: Nothing
                   }

    render self =
      R.div_
      [ datasetSelector { datasets: self.props.config.datasets
                        , onDatasetSelected: update self <<< DatasetSelected
                        }

      , foldMap (\ds -> dirBrowser
                       { dir: ds.mountPoint
                       , onFileSelected: update self <<< FileSelected
                       , onDirSelected: update self <<< DirSelected
                       } ) self.state.selectedDataset
 

      , foldMap (\file -> fileVersionSelector
                         { file
                         , onVersionSelected: update self <<< VersionSelected 
                         }) self.state.selectedFile

        
      , foldMap (uncurry2 (\file version -> fileAction { file, version }))
                $ (tuple2 <$> self.state.selectedFile
                          <*> self.state.selectedVersion)
      ]
