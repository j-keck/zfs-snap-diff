module ZSD.Model.DateRange
       -- ( DateRange(..)
       -- , lastNDays
       -- , dayCount
       -- , slide
       -- , adjustFrom
       -- , adjustTo
       -- )
       where

import Prelude

import Data.Array ((..))
import Control.Monad.Except (except)
import Data.Bifunctor (lmap)
import Data.Date (Date)
import Data.Date as Date
import Data.DateTime as DT
import Data.Enum (class BoundedEnum, fromEnum, toEnum)
import Data.Formatter.DateTime (FormatterCommand(..), format, unformatDateTime)
import Data.Int (floor)
import Data.List as L
import Data.List.NonEmpty as LNE
import Data.Maybe (fromMaybe)
import Data.Newtype (class Newtype, over, unwrap)
import Data.Time.Duration (Days(..), Milliseconds)
import Debug.Trace (spy)
import Effect (Effect)
import Effect.Now as Effect
import Foreign (F, ForeignError(..))
import Simple.JSON (class ReadForeign, class WriteForeign)
import Simple.JSON as F

newtype DateRange = DateRange
  { from :: Date
  , to   :: Date
  }


lastNDays :: Days -> Effect DateRange
lastNDays days = do
  now <- Effect.nowDate
  let from = adjustDate (over Days negate days) now
  pure $ DateRange { from, to: now }


dayCount :: DateRange -> Int
dayCount (DateRange { from, to }) =
  let (days :: Milliseconds) = Date.diff to from
  in add 1 <<< floor $ (unwrap days) / 86400000.0


slide :: Days -> DateRange -> DateRange
slide days d = mapDates (adjustDate $ over Days negate days)
                        (const $ (unwrap >>> _.from >>> adjustDate (Days (-1.0))) d)
                        d


adjustFrom :: Days -> DateRange -> DateRange
adjustFrom days = mapDates (adjustDate days) identity


adjustTo :: Days -> DateRange -> DateRange
adjustTo days = mapDates identity (adjustDate days)


mapDates :: (Date -> Date) -> (Date -> Date) -> DateRange -> DateRange
mapDates updFrom updTo = over DateRange (\rec -> rec { from = updFrom rec.from
                                                     , to = updTo rec.to
                                                     })


adjustDate :: Days -> Date -> Date
adjustDate days d = fromMaybe d $ Date.adjust days d


instance showDateRange :: Show DateRange where
  show (DateRange { from, to }) =
    "DateRange { from: " <> showDate from <> ", to: " <> showDate to <> "}"
    where showDate d = s DT.year d  <> "-" <> s DT.month d <> "-" <> s DT.day d
          s :: forall a b. BoundedEnum b => (a -> b) -> a -> String
          s get = get >>> fromEnum >>> show


derive newtype instance eqDateRange :: Eq DateRange
derive instance newtypeDateRange :: Newtype DateRange _

instance readForgeignDateRange :: ReadForeign DateRange where
  readImpl f = F.readImpl f >>= toDateRange
    where toDateRange :: { from :: String, to :: String } -> F DateRange
          toDateRange { from: fromS, to: toS } = do
            from <- parseDate fromS
            to <- parseDate toS
            pure $ DateRange { from, to }

          parseDate s = except $ wrapLeft $ DT.date <$> unformatDateTime "YYYY-MM-DD" s
          wrapLeft = lmap (ForeignError >>> LNE.singleton)


instance writeForeignDateRange :: WriteForeign DateRange where
  writeImpl (DateRange { from, to }) = F.writeImpl { from: toS from, to: toS to }
    where toS :: Date -> String
          toS date = let dt = DT.DateTime date bottom
                         dash = Placeholder "-"
                         fmt = L.fromFoldable [YearFull, dash, MonthTwoDigits, dash, DayOfMonthTwoDigits]
                     in format fmt dt

instance semigroupDateRange :: Semigroup DateRange where
  append (DateRange { from: f1, to: t1 }) (DateRange { from: f2, to: t2 })
    | f1 < f2   = DateRange { from: f1, to: t2 }
    | otherwise = DateRange { from: f2, to: t1 }
