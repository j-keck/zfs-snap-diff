module ZSD.View.BrowseSnapshots where


import Prelude

import Data.Array as A
import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.String as S
import Data.Tuple.Nested (tuple2, uncurry2)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import ZSD.Components.DatasetSelector (datasetSelector)
import ZSD.Components.DirBrowser (dirBrowser)
import ZSD.Components.FileActions (fileAction)
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.Snapshot (Snapshots, Snapshot)
import ZSD.Model.Snapshot as Snapshots
import ZSD.Ops (unsafeFromJust)
import ZSD.Views.BrowseSnapshots.SnapshotSelector (snapshotSelector)


type Props = { config :: Config }

type State =
  { selectedDataset  :: Maybe Dataset
  , selectedSnapshot :: Maybe Snapshot
  , selectedFile     :: Maybe FSEntry
  , selectedDir      :: Maybe FSEntry
  } 


data Command =
    DatasetSelected Dataset

update :: (React.Self Props State) -> Command -> Effect Unit
update self = case _ of
  DatasetSelected ds -> do
    self.setState _ { selectedDataset = Just ds }


browseSnapshots :: Props -> JSX
browseSnapshots = make component { initialState, render }
  where

    component :: Component Props
    component = createComponent "BrowseSnapshots"

    initialState = { selectedDataset: Nothing
                   , selectedSnapshot: Nothing
                   , selectedFile: Nothing
                   , selectedDir: Nothing
                   }

    render self =
      R.div_
      [ datasetSelector { datasets: self.props.config.datasets
                        , onDatasetSelected: update self <<< DatasetSelected
                        }

      , foldMap (\dataset -> snapshotSelector
                              { dataset 
                              , onSnapshotSelected: \snap -> self.setState _ { selectedSnapshot = Just snap, selectedFile = Nothing, selectedDir = Nothing }
                              }) self.state.selectedDataset
        
      , foldMap (\snapshot -> dirBrowser
                              { dir: snapshot.dir 
                              , onFileSelected: \file -> self.setState _ { selectedFile = Just file }
                              , onDirSelected: \dir -> self.setState _ { selectedDir = Just dir, selectedFile = Nothing }
                              }) self.state.selectedSnapshot

      , foldMap (uncurry2 (\file snapshot ->
                  -- FIXME: cleanup: update the file path in the dataset
                  let snapPathElements = S.split (S.Pattern "/") snapshot.dir.path
                      filePathElements = S.split (S.Pattern "/") file.path
                      relPath = S.joinWith "/" $ A.drop (A.length snapPathElements) filePathElements
                      dsPath = (unsafeFromJust self.state.selectedDataset).mountPoint.path
                      file' = file { path = dsPath <> "/" <> relPath } 
                      version = BackupVersion {file, snapshot}
                  in fileAction { file: file', version }))
          (tuple2 <$> self.state.selectedFile
           <*> self.state.selectedSnapshot)
                  
      ]
        
 
