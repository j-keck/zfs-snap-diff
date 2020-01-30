module ZSD.View.BrowseSnapshots where


import Prelude

import Data.Array as A
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.String as S
import Data.Tuple.Nested (tuple2, uncurry2)
import Effect (Effect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import ZSD.Fragments.DatasetSelector (datasetSelector)
import ZSD.Fragments.DirBrowser (dirBrowser)
import ZSD.Fragments.FileActions (fileAction)
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.Snapshot (Snapshot)
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
  | SnapshotSelected Snapshot
  | FileSelected FSEntry
  | DirSelected FSEntry

update :: (React.Self Props State) -> Command -> Effect Unit
update self = case _ of
  DatasetSelected ds ->
    self.setState _ { selectedDataset = Just ds
                    , selectedSnapshot = Nothing
                    , selectedFile = Nothing
                    , selectedDir = Nothing
                    }

  SnapshotSelected snap ->
    self.setState _ { selectedSnapshot = Just snap
                    , selectedFile = Nothing
                    , selectedDir = Nothing
                    }

  FileSelected fh ->
    self.setState _ { selectedFile = Just fh }

  DirSelected fh ->
    self.setState _ { selectedDir = Just fh
                    , selectedFile = Nothing
                    }



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
                              , onSnapshotSelected: update self <<< SnapshotSelected
                              }) self.state.selectedDataset

      , foldMap (uncurry2 (\ds snapshot -> dirBrowser
                              { ds
                              , root: snapshot.dir
                              , onFileSelected: update self <<< FileSelected
                              , onDirSelected: update self <<< DirSelected
                              })) (tuple2 <$> self.state.selectedDataset <*> self.state.selectedSnapshot)

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


