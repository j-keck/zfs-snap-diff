module ZSD.Views.BrowseFilesystem.FileVersionSelector where

import Prelude
import ZSD.Model.DateRange
import ZSD.Ops

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
import React.Basic.DOM.Textf as TF
import ZSD.Component.TableX (tableX)
import ZSD.Components.Panel (panel)
import ZSD.Formatter as Formatter
import ZSD.Model.DateRange as DateRange
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Model.ScanResult (ScanResult(..))
import ZSD.Model.ScanResult as ScanResult



type Props =
  { file :: FSEntry
  , onVersionSelected :: FileVersion -> Effect Unit
  }

type State = { versions :: Array FileVersion, selectedIdx :: Int, scanResults :: Array ScanResult }

data Action =
    DidMount
  | Scan (Maybe DateRange) Action
  | SelectVersionByIdx Int


data SelectedVersion =
    Actual FileVersion
  | Backup FileVersion Int
 
update :: React.Self Props State -> Action -> Effect Unit
update self = case _ of
  DidMount -> do
    self.setState _ { versions = [ ActualVersion self.props.file ] }
    update self $ Scan Nothing (SelectVersionByIdx 0)

        
  Scan (Just range) next -> launchAff_ $ do
    scanResult <- unsafePartial $ fromRight <$> ScanResult.fetch self.props.file range
    liftEffect $ do
      state <- readState self
      let versions = A.concat [state.versions, (unwrap scanResult).fileVersions]
      self.setStateThen (const $ state { scanResults = state.scanResults `A.snoc` scanResult
                                       , versions = versions })
                               $ update self next

  Scan Nothing next -> do
    state <- readState self
    dateRange <- maybe (DateRange.lastNDays 1)
                       (unwrap >>> _.dateRange >>> DateRange.slide (Date.Days $ -7.0) >>> pure)
                       (A.last state.scanResults )
    update self $ Scan (Just dateRange) next
        
        
  SelectVersionByIdx idx -> do
    state <- readState self
    case A.index state.versions idx of
      Just next -> self.setStateThen _ { selectedIdx = idx } $ self.props.onVersionSelected next
      Nothing -> guard (hasOlderVersions state) $ update self $ Scan Nothing (SelectVersionByIdx idx)


fileVersionSelector :: Props -> JSX
fileVersionSelector = make component { initialState, didMount, render } 

  where

     component :: Component Props
     component = createComponent "FileVersionSelector"

     initialState = { versions: [], selectedIdx: 0 , scanResults: [] }

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
                                       $ update self (SelectVersionByIdx (self.state.selectedIdx + 1))
                   , children:
                     [ R.span { className: "fas fa-backward p-1" }
                     , R.text "Older"
                     ]
                   }
                 , R.button
                   { className: "btn btn-secondary" <> guard (not $ hasNewerVersions self.state) " disabled"
                   , title: "Select the successor version"
                   , onClick: capture_ $ guard (hasNewerVersions self.state)
                                       $ update self (SelectVersionByIdx (self.state.selectedIdx - 1))
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
               self.setState _ { selectedIdx = idx }
               self.props.onVersionSelected v
           , activeIdx: Just self.state.selectedIdx
           }

       , footer: 
           R.div
           { className: "text-muted small text-center"
           , children:
             [ flip foldMap (A.last self.state.scanResults)
                 \(ScanResult { snapsScanned, dateRange: (DateRange range), scanDuration }) ->
                   R.div_
                   [ TF.textf
                     [ TF.text { style: TF.textUnderline } "This scan: "
                     , TF.text' "Scanned ", TF.int { style: TF.fontBolder } snapsScanned, TF.text' " snapshots in "
                     , TF.text { style: TF.fontBolder } (Formatter.duration scanDuration), TF.text' ". "
                     , TF.text' "Date range: "
                     , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.from, TF.text' " - "
                     , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.to, TF.text' "."
                     ]
                   ]
             , flip foldMap (foldlSemigroup self.state.scanResults)
                 \(ScanResult { snapsScanned, snapsToScan, snapsFileMissing, dateRange: (DateRange range), scanDuration }) ->
                   R.div_ 
                   [ TF.textf
                     [ TF.text { style: TF.textUnderline } "Overall: "
                     , TF.text' "Scanned ", TF.int { style: TF.fontBolder } snapsScanned, TF.text' " snapshots in "
                     , TF.text { style: TF.fontBolder } (Formatter.duration scanDuration), TF.text' ". "
                     , TF.text' "Date range: "
                     , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.from, TF.text' " - "
                     , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.to, TF.text' "."
                     ]                     
                  ] <>
                  R.div_
                  [ TF.textf
                    [ TF.int { style: TF.fontBolder } snapsToScan, TF.text' " snapshots to scan."
                    , TF.text' " In ", TF.int { style: TF.fontBolder } snapsFileMissing, TF.text' " snapshots was the file not found."
                    ]
                  ]
              ]
           }
       }



       


hasNewerVersions :: State -> Boolean
hasNewerVersions state = state.selectedIdx /= 0

hasOlderVersions :: State -> Boolean 
hasOlderVersions state = state.selectedIdx < (A.length state.versions) - 1 ||
                         maybe false (unwrap >>> _.snapsToScan >>> (/=) 0) (A.last state.scanResults)

