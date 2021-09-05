-- | BrowseSnapshots lists all available snapshots for a dataset
module ZSD.Views.BrowseSnapshots where

import Prelude
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.Either (either)
import Data.Tuple.Nested (tuple2, tuple3, uncurry2, uncurry3)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, make)
import React.Basic.Classic as React
import React.Basic.DOM as R
import ZSD.Fragments.DatasetSelector (datasetSelector)
import ZSD.Fragments.DirBrowser (dirBrowser)
import ZSD.Fragments.FileActions (fileAction)
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset, Datasets)
import ZSD.Model.Dataset as Dataset
import ZSD.Model.FH (FH, From(..), To(..), switchMountPoint)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Views.BrowseSnapshots.SnapshotSelector (snapshotSelector)
import ZSD.Components.Spinner as Spinner
import ZSD.Views.Messages as Messages

type Props
  = { config :: Config
    , activeDataset :: Maybe Dataset
    , onDatasetSelected :: Dataset -> Effect Unit
    }

type State
  = { datasets :: Datasets
    , selectedDataset :: Maybe Dataset
    , selectedSnapshot :: Maybe Snapshot
    , selectedFile :: Maybe FH
    , selectedDir :: Maybe FH
    }

data Command
  = DatasetSelected Dataset
  | FetchDatasets
  | SnapshotSelected Snapshot
  | FileSelected FH
  | DirSelected FH

update :: (React.Self Props State) -> Command -> Effect Unit
update self = case _ of
  FetchDatasets ->
    Spinner.display
      *> launchAff_
          ( Dataset.fetch
              >>= either Messages.appError (\ds -> self.setState _ { datasets = ds } *> Spinner.remove)
              >>> liftEffect
          )
  DatasetSelected ds -> do
    -- FIXME: when a snapshot are created from the 'DatasetSelector' fragment,
    -- the new snapshot was not shown. Resetting the 'selectedDataset' triggers
    -- a reload of the component. FIX: include a event notification
    self.setState
      _
        { selectedDataset = Nothing
        , selectedSnapshot = Nothing
        , selectedFile = Nothing
        , selectedDir = Nothing
        }
    self.setState _ { selectedDataset = Just ds }
    self.props.onDatasetSelected ds
  SnapshotSelected snap ->
    self.setState
      _
        { selectedSnapshot = Just snap
        , selectedFile =
          do
            oldSnap <- self.state.selectedSnapshot
            file <- self.state.selectedFile
            pure $ switchMountPoint (From oldSnap.mountPoint) (To snap.mountPoint) file
        }
  FileSelected fh -> self.setState _ { selectedFile = Just fh }
  DirSelected fh ->
    self.setState
      _
        { selectedDir = Just fh
        , selectedFile = Nothing
        }

browseSnapshots :: Props -> JSX
browseSnapshots = make component { initialState, didMount, render }
  where
  component :: Component Props
  component = createComponent "BrowseSnapshots"

  initialState =
    { datasets: []
    , selectedDataset: Nothing
    , selectedSnapshot: Nothing
    , selectedFile: Nothing
    , selectedDir: Nothing
    }

  didMount self = self.setStateThen _ { selectedDataset = self.props.activeDataset } $ update self FetchDatasets

  render self =
    R.div_
      [ datasetSelector
          { datasets: self.state.datasets
          , activeDataset: self.props.activeDataset
          , onDatasetSelected: update self <<< DatasetSelected
          , snapshotNameTemplate: self.props.config.snapshotNameTemplate
          }
      -- snapshot selector
      , foldMap
          ( \dataset ->
              snapshotSelector
                { dataset
                , onSnapshotSelected: update self <<< SnapshotSelected
                , onDatasetChanges: update self FetchDatasets
                }
          )
          self.state.selectedDataset
      -- dir browser
      , foldMap
          ( uncurry2
              ( \ds snapshot ->
                  dirBrowser
                    { ds
                    , snapshot: Just snapshot
                    , onFileSelected: update self <<< FileSelected
                    , onDirSelected: update self <<< DirSelected
                    }
              )
          )
          (tuple2 <$> self.state.selectedDataset <*> self.state.selectedSnapshot)
      -- file actions
      , foldMap
          ( uncurry3
              ( \ds snapshot file ->
                  let
                    current = switchMountPoint (From snapshot.mountPoint) (To ds.mountPoint) file

                    version = BackupVersion { current, backup: file, snapshot }
                  in
                    fileAction { file, version }
              )
          )
          ( tuple3 <$> self.state.selectedDataset
              <*> self.state.selectedSnapshot
              <*> self.state.selectedFile
          )
      ]
