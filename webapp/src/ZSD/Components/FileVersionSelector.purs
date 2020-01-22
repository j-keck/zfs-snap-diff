module ZSD.Components.FileVersionSelector where

import Prelude

import Data.Array as A
import Data.Either (fromRight)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), maybe)
import Data.Monoid (guard)
import Data.Newtype (unwrap)
import Data.Time.Duration as Date
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, fragment, make, readState)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Component.TableX (tableX)
import ZSD.Components.Panel (panel)
import ZSD.Formatter as Formatter
import ZSD.Model.DateRange (DateRange)
import ZSD.Model.DateRange as DateRange
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.ScanResult (ScanResult)
import ZSD.Model.ScanResult as ScanResult



type Props =
  { file :: FSEntry
  , onVersionSelected :: FileVersion -> Effect Unit
  }

type State = { versions :: Array FileVersion, idx :: Int, scanResult :: Maybe ScanResult }

data Action =
    DidMount
  | Scan (Maybe DateRange) Action
  | SelectVersionByIdx Int

 
update :: React.Self Props State -> Action -> Effect Unit
update self = case _ of
  DidMount -> do
    self.setState _ { versions = [ ActualVersion self.props.file ] }
    update self $ Scan Nothing (SelectVersionByIdx 0)

        
  Scan (Just range) next -> launchAff_ $ do
    scanResult <- unsafePartial $ fromRight <$> ScanResult.fetch self.props.file range
    liftEffect $ do
      state <- readState self
      let versions = A.concat [state.versions, scanResult.fileVersions]
      self.setStateThen (const $ state { scanResult = Just scanResult
                                       , versions = versions })
                               $ update self next

  Scan Nothing next -> do
    state <- readState self
    dateRange <- maybe (DateRange.lastNDays 1)
                       (_.scannedDateRange >>> DateRange.slide (Date.Days $ -7.0) >>> pure)
                       state.scanResult
    update self $ Scan (Just dateRange) next
        
        
  SelectVersionByIdx idx -> do
    state <- readState self
    case A.index state.versions idx of
      Just next -> self.setStateThen _ { idx = idx } $ self.props.onVersionSelected next
      Nothing -> guard (hasOlderVersions state) $ update self $ Scan Nothing (SelectVersionByIdx idx)


fileVersionSelector :: Props -> JSX
fileVersionSelector = make component { initialState, didMount, render } 

  where

     component :: Component Props
     component = createComponent "FileVersionSelector"

     initialState = { versions: [], idx: -1 , scanResult: Nothing }

     didMount self = update self DidMount

     render self =
       panel 
       { header: fragment 
         [ R.text $ "Versions for file: " <> self.props.file.name
         , R.span
           { className: "float-right" 
           , children:
             [ R.div
               { className: "btn-group"
               , children:
                 [ R.button
                   { className: "btn btn-secondary" <> guard  (not $ hasOlderVersions self.state) " disabled"
                   , title: "Select the previous version"
                   , onClick: capture_ $ guard (hasOlderVersions self.state)
                                       $ update self (SelectVersionByIdx (self.state.idx + 1))
                   , children:
                     [ R.span { className: "fas fa-backward p-1" }
                     , R.text "Older"
                     ]
                   }
                 , R.button
                   { className: "btn btn-secondary" <> guard (not $ hasNewerVersions self.state) " disabled"
                   , title: "Select the successor version"
                   , onClick: capture_ $ guard (hasNewerVersions self.state)
                                       $ update self (SelectVersionByIdx (self.state.idx - 1))
                   , children:
                     [ R.text "Newer"
                     , R.span { className: "fas fa-forward p-1" }
                     ]
                   }
                 ]
               }
             ]
           }
         ]
       , body: \hidePanelBodyFn ->
           tableX
           { header: ["File modification time", "Snapshot Created", "Snapshot Name"]
           , rows: self.state.versions
           , mkRow: case _ of
               ActualVersion f -> [ R.text $ "Actual version", R.text "-", R.text "-" ]
               BackupVersion v -> [ R.text $ Formatter.dateTime v.file.modTime
                                  , R.text $ Formatter.dateTime v.snapshot.created
                                  , R.text v.snapshot.name
                                  ]
           , onRowSelected: \(Tuple idx v) -> do
               hidePanelBodyFn
               self.setState _ { idx = idx }
               self.props.onVersionSelected v
           , activeIdx: Just self.state.idx
           }

       , footer: flip foldMap self.state.scanResult $ \sr -> 
           R.div
           { className: "text-muted small text-center"
           , children:
             [ R.text "Scanned "
             , R.b_ [ R.text (show $ DateRange.dayCount sr.scannedDateRange) ]
             , R.text " days (snapshots between "
             , R.text (Formatter.date (unwrap sr.scannedDateRange).from)
             , R.text " and "
             , R.text (Formatter.date (unwrap sr.scannedDateRange).to)
             , R.text ")."
             , R.text " Scan duration: "
             , R.b_ [ R.text (Formatter.duration sr.scanDuration) ]
             ]
           } <> stats sr
       }


     stats sr =
       R.div
       { className: "text-muted small text-center"
       , children:
         [ mkStat sr.scannedSnapshots " snapshots scanned, " "Scanned snapshots in the last scan"
         , mkStat sr.skippedSnapshots " snapshots skipped, " "Skipped snapshots are already scanned snapshots"
         , mkStat sr.snapshotsToScan " snapshots to scan." ""
         , R.text " In "
         , mkStat sr.fileMissingSnapshots " snapshots the file was missing." ""
         ]
       }
       
     mkStat n text title =
       fragment
       [ R.b_ [ R.text $ show n ]
       , R.span
         { title
         , children: [ R.text text ]
         }
       ]



hasNewerVersions :: State -> Boolean
hasNewerVersions state = state.idx /= 0

hasOlderVersions :: State -> Boolean
hasOlderVersions state = state.idx < (A.length state.versions) - 1 ||
                         maybe false (\sr -> sr.snapshotsToScan > 0) state.scanResult

