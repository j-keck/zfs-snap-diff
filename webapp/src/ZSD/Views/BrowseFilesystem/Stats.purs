module ZSD.Views.BrowseFilesystem.Stats where

import Prelude

import Data.Array as A
import Data.Traversable (foldMap)
import React.Basic (JSX, createComponent, makeStateless)
import React.Basic.DOM as R
import React.Basic.DOM.Textf as TF
import ZSD.Model.DateRange (DateRange(..))
import ZSD.Model.FileVersion (ScanResult(..))
import ZSD.Utils.Formatter as Formatter
import ZSD.Utils.Ops (foldlSemigroup)

type Props = Array ScanResult

stats :: Props -> JSX
stats = makeStateless component \props ->
  R.div
  { className: "text-muted small text-center"
  , children: let fmtDate = TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy], style: TF.fontBolder } in
    [ flip foldMap (A.last props)
        \(ScanResult { snapsScanned, dateRange: (DateRange range), scanDuration }) ->
          R.div_
          [ TF.textf
            [ TF.text { style: TF.textUnderline } "This scan: "
            , TF.text' "Scanned ", TF.int { style: TF.fontBolder } snapsScanned, TF.text' " snapshots in "
            , TF.text { style: TF.fontBolder } (Formatter.duration scanDuration), TF.text' ". "
            , TF.text' "Date range: "
            , fmtDate range.from, TF.text' " - "
            , fmtDate range.to, TF.text' "."
            ]
          ]
    , flip foldMap (foldlSemigroup props)
        \(ScanResult { snapsScanned, snapsToScan, snapsFileMissing, dateRange: (DateRange range), scanDuration }) ->
          R.div_
          [ TF.textf
            [ TF.text { style: TF.textUnderline } "Overall: "
            , TF.text' "Scanned ", TF.int { style: TF.fontBolder } snapsScanned, TF.text' " snapshots in "
            , TF.text { style: TF.fontBolder } (Formatter.duration scanDuration), TF.text' ". "
            , TF.text' "Date range: "
            , fmtDate range.from, TF.text' " - "
            , fmtDate range.to, TF.text' "."
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

  where
    component = createComponent "Stats"
