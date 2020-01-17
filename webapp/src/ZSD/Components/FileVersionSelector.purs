module ZSD.Components.FileVersionSelector where

import Prelude

import Data.Array as A
import Data.Either (fromRight)
import Data.Maybe (Maybe(..), fromJust, maybe)
import Data.Monoid (guard)
import Data.Newtype (unwrap)
import Data.Time.Duration as Date
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, fragment, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Components.LogLifecycles (logLifecycles)
import React.Basic.DOM.Events (capture_)
import ZSD.Component.Table (table)
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

type State = { versions :: Array FileVersion, scanResult :: Maybe ScanResult }

data Action =
    Scan DateRange
  | ScanMore

hasSnapshotsToScan :: Maybe ScanResult -> Boolean
hasSnapshotsToScan = case _ of
  Just sr -> sr.snapshotsToScan > 0
  Nothing -> false


update :: React.Self Props State -> Action -> Effect Unit
update self = case _ of
  Scan dateRange -> launchAff_ $ do
    scanResult <- unsafePartial $ fromRight <$> ScanResult.fetch self.props.file dateRange
    let versions = A.concat [self.state.versions, scanResult.fileVersions]
    liftEffect $ self.setState \s -> s { scanResult = Just scanResult
                                       , versions = versions }
    
  ScanMore -> guard (hasSnapshotsToScan self.state.scanResult) do
    let dateRange = unsafePartial $ (fromJust self.state.scanResult).scannedDateRange
    update self $ Scan (DateRange.slide (Date.Days $ -7.0) dateRange)




fileVersionSelector :: Props -> JSX
fileVersionSelector props = logLifecycles $ make component { initialState, didMount, render } props

  where

     component :: Component Props
     component = createComponent "FileVersionSelector"

     initialState = { versions: [ActualVersion props.file], scanResult: Nothing }

     didMount self = do
       log "FileVersionSelector - didMount"
       DateRange.lastNDays 1 >>= update self <<< Scan


     render self =
       panel
       { title: "Versions for file: " <> self.props.file.name
       , body:
         \hidePanelBodyFn -> fragment
           [ table
             { header: ["File modification time", "Snapshot Created", "Snapshot Name"]
             , rows: self.state.versions
             , mkRow: case _ of
                 ActualVersion f -> [ R.text $ "Actual version", R.text "-", R.text "-" ]
                 BackupVersion v -> [ R.text $ Formatter.dateTime v.file.modTime
                                    , R.text $ Formatter.dateTime v.snapshot.created
                                    , R.text v.snapshot.name
                                    ]
             , onRowSelected: \v -> do
                 hidePanelBodyFn
                 self.props.onVersionSelected v
             }
           , flip (maybe mempty) self.state.scanResult $ \scanResult ->
             R.div_
             [
               R.button { className: "btn btn-secondary btn-block mb-1" <> guard (not $ hasSnapshotsToScan self.state.scanResult) " disabled"
                        , onClick: capture_ $ update self ScanMore
                        , children: [ R.text "Scan older snapshots for other file versions"]
                        }
             , R.div
               { className: "text-muted small text-center"
               , children:
                 [ R.text "Scanned "
                 , R.b_ [ R.text (show $ DateRange.dayCount scanResult.scannedDateRange) ]
                 , R.text " days (snapshots between "
                 , R.text (Formatter.date (unwrap scanResult.scannedDateRange).from)
                 , R.text " and "
                 , R.text (Formatter.date (unwrap scanResult.scannedDateRange).to)
                 , R.text ")."
                 , R.text " Scan duration: "
                 , R.b_ [ R.text (Formatter.duration scanResult.scanDuration) ]
                 ]
               }
             , stats scanResult
             ]
           ]
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

