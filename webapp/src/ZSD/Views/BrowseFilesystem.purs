module ZSD.Views.BrowseFilesystem where

import Prelude
import Data.Either (fromRight)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.Tuple.Nested (tuple3, uncurry3)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Components.LogLifecycles (logLifecycles)
import ZSD.Components.DatasetSelector (datasetSelector)
import ZSD.Components.DirBrowser (dirBrowser)
import ZSD.Components.FileActions (fileAction)
import ZSD.Components.FileVersionSelector (fileVersionSelector)
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion)
import ZSD.Model.MimeType (MimeType)
import ZSD.Model.MimeType as MimeType


type Props = { config :: Config }
type State = { selectedDataset      :: Maybe Dataset
             , selectedFile         :: Maybe FSEntry
             , selectedFileMimeType :: Maybe MimeType
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
                    , selectedFileMimeType = Nothing
                    , selectedVersion = Nothing
                    }

  FileSelected f -> launchAff_ $ do
    mimeType <- unsafePartial $ fromRight <$> MimeType.fetch f
    liftEffect $ self.setState _ { selectedFile = Just f
                                 , selectedFileMimeType = Just mimeType
                                 , selectedVersion = Nothing
                                 }

  DirSelected d ->
    self.setState _ { selectedFile = Nothing
                    , selectedFileMimeType = Nothing
                    , selectedVersion = Nothing
                    }

  VersionSelected v ->
    self.setState _ { selectedVersion = Just v }




browseFilesystem :: Props -> JSX
browseFilesystem = logLifecycles <<< make component { initialState, render }

  where

    component :: Component Props
    component = createComponent "BrowseFilesystem"

    initialState = { selectedDataset: Nothing
                   , selectedFile: Nothing
                   , selectedFileMimeType: Nothing
                   , selectedVersion: Nothing
                   }

    render self =
      R.div_
      [ datasetSelector { datasets: self.props.config.datasets
                        , onDatasetSelected: update self <<< DatasetSelected
                        }

      , foldMap (\ds -> dirBrowser
                       { dataset: ds
                       , onFileSelected: update self <<< FileSelected
                       , onDirSelected: update self <<< DirSelected
                       } ) self.state.selectedDataset

      , foldMap (\file -> fileVersionSelector
                         { file
                         , onVersionSelected: update self <<< VersionSelected
                         }) self.state.selectedFile

      , foldMap (uncurry3 (\file mimeType version -> fileAction { file, mimeType, version }))
                $ (tuple3 <$> self.state.selectedFile
                          <*> self.state.selectedFileMimeType
                          <*> self.state.selectedVersion)
      ]
