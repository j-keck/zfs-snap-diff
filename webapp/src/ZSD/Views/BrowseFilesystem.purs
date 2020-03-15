module ZSD.Views.BrowseFilesystem where

import Prelude
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.Monoid (guard)
import Data.Either (either)
import Data.Tuple.Nested (tuple2, uncurry2)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import ZSD.Fragments.DatasetSelector (datasetSelector)
import ZSD.Fragments.DirBrowser (dirBrowser)
import ZSD.Fragments.FileActions (fileAction)
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset, Datasets)
import ZSD.Model.Dataset as Dataset
import ZSD.Model.FH (FH)
import ZSD.Model.FileVersion (FileVersion)
import ZSD.Views.BrowseFilesystem.FileVersionSelector (fileVersionSelector)
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
    , selectedFile :: Maybe FH
    , selectedVersion :: Maybe FileVersion
    }

data Command
  = DatasetSelected Dataset
  | FileSelected FH
  | DirSelected FH
  | VersionSelected FileVersion

update :: (React.Self Props State) -> Command -> Effect Unit
update self = case _ of
  DatasetSelected ds ->
    guard (Just ds /= self.state.selectedDataset) do
      self.setState
        _
          { selectedDataset = Just ds
          , selectedFile = Nothing
          , selectedVersion = Nothing
          }
      self.props.onDatasetSelected ds
  FileSelected f ->
    self.setState
      _
        { selectedFile = Just f
        , selectedVersion = Nothing
        }
  DirSelected d ->
    self.setState
      _
        { selectedFile = Nothing
        , selectedVersion = Nothing
        }
  VersionSelected v -> self.setState _ { selectedVersion = Just v }

browseFilesystem :: Props -> JSX
browseFilesystem props = make component { initialState, didMount, render } props
  where
  component :: Component Props
  component = createComponent "BrowseFilesystem"

  initialState =
    { datasets: []
    , selectedDataset: Nothing
    , selectedFile: Nothing
    , selectedVersion: Nothing
    }

  --didMount self = log "DID_MOUNT" *> self.setState _ { selectedDataset = self.props.activeDataset }
  didMount self =
    Spinner.display
      *> launchAff_
          ( Dataset.fetch
              >>= either Messages.appError
                  ( \ds ->
                      self.setState
                        _
                          { datasets = ds
                          , selectedDataset = self.props.activeDataset
                          }
                        *> Spinner.remove
                  )
              >>> liftEffect
          )

  render self =
    R.div_
      [ datasetSelector
          { datasets: self.state.datasets
          , activeDataset: self.props.activeDataset
          , onDatasetSelected: update self <<< DatasetSelected
          , snapshotNameTemplate: self.props.config.snapshotNameTemplate
          }
      -- dir browser
      , foldMap
          ( \ds ->
              dirBrowser
                { ds
                , snapshot: Nothing
                , onFileSelected: update self <<< FileSelected
                , onDirSelected: update self <<< DirSelected
                }
          )
          self.state.selectedDataset
      -- file version selector
      , foldMap
          ( \file ->
              fileVersionSelector
                { file
                , onVersionSelected: update self <<< VersionSelected
                , daysToScan: self.props.config.daysToScan
                }
          )
          self.state.selectedFile
      -- file actions
      , foldMap (uncurry2 (\file version -> fileAction { file, version }))
          $ ( tuple2 <$> self.state.selectedFile
                <*> self.state.selectedVersion
            )
      ]
