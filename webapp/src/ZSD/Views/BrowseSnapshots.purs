-- | BrowseSnapshots lists all available snapshots for a dataset
module ZSD.Views.BrowseSnapshots where

import Prelude

import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.Tuple.Nested (tuple2, tuple3, uncurry2, uncurry3)
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import ZSD.Fragments.DatasetSelector (datasetSelector)
import ZSD.Fragments.DirBrowser (dirBrowser)
import ZSD.Fragments.FileActions (fileAction)
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FH (FH, From(..), To(..), switchMountPoint)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Views.BrowseSnapshots.SnapshotSelector (snapshotSelector)


type Props =
  { config            :: Config
  , activeDataset     :: Maybe Dataset
  , onDatasetSelected :: Dataset -> Effect Unit
  }

type State =
  { selectedDataset  :: Maybe Dataset
  , selectedSnapshot :: Maybe Snapshot
  , selectedFile     :: Maybe FH
  , selectedDir      :: Maybe FH
  }


data Command =
    DatasetSelected Dataset
  | SnapshotSelected Snapshot
  | FileSelected FH
  | DirSelected FH


update :: (React.Self Props State) -> Command -> Effect Unit
update self = case _ of
  DatasetSelected ds -> do
    -- FIXME: when a snapshot are created from the 'DatasetSelector' fragment,
    -- the new snapshot was not shown. Resetting the 'selectedDataset' triggers
    -- a reload of the component. FIX: include a event notification
    self.setState _ { selectedDataset = Nothing
                    , selectedSnapshot = Nothing
                    , selectedFile = Nothing
                    , selectedDir = Nothing
                    }
    self.setState _ { selectedDataset = Just ds }
    self.props.onDatasetSelected ds

  SnapshotSelected snap ->
    self.setState _ { selectedSnapshot = Just snap
                    , -- when switching the snapshot, the path of the selected
                      -- file must be changed
                      selectedFile = do
                        oldSnap <- self.state.selectedSnapshot
                        file <- self.state.selectedFile
                        pure $ switchMountPoint (From oldSnap.mountPoint) (To snap.mountPoint) file
                    }

  FileSelected fh ->
    self.setState _ { selectedFile = Just fh }


  DirSelected fh ->
    self.setState _ { selectedDir = Just fh
                    , selectedFile = Nothing
                    }



browseSnapshots :: Props -> JSX
browseSnapshots = make component { initialState, didMount, render }
  where

    component :: Component Props
    component = createComponent "BrowseSnapshots"

    initialState = { selectedDataset: Nothing
                   , selectedSnapshot: Nothing
                   , selectedFile: Nothing
                   , selectedDir: Nothing
                   }

    didMount self = self.setState _ { selectedDataset = self.props.activeDataset }

    render self =
      R.div_
      [
        -- dataset selector
        datasetSelector { datasets: self.props.config.datasets
                        , activeDataset: self.props.activeDataset
                        , onDatasetSelected: update self <<< DatasetSelected
                        , snapshotNameTemplate: self.props.config.snapshotNameTemplate
                        }

        -- snapshot selector
      , foldMap (\dataset -> snapshotSelector
                              { dataset
                              , onSnapshotSelected: update self <<< SnapshotSelected
                              }) self.state.selectedDataset

        -- dir browser
      , foldMap (uncurry2 (\ds snapshot -> dirBrowser
                              { ds
                              , snapshot: Just snapshot
                              , onFileSelected: update self <<< FileSelected
                              , onDirSelected: update self <<< DirSelected
                              })) (tuple2 <$> self.state.selectedDataset <*> self.state.selectedSnapshot)

        -- file actions
      , foldMap (uncurry3 (\ds snapshot file ->
                            let actual = switchMountPoint (From snapshot.mountPoint) (To ds.mountPoint) file
                                version = BackupVersion { actual, backup: file, snapshot }
                            in fileAction { file, version }
                          ))
          (tuple3 <$> self.state.selectedDataset
                  <*> self.state.selectedSnapshot
                  <*> self.state.selectedFile)

      ]
