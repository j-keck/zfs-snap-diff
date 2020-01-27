module ZSD.Model.DateRange
       -- ( DateRange(..)
       -- , lastNDays
       -- , dayCount
       -- , slide
       -- , adjustFrom
       -- , adjustTo
       -- )
       where

import Data.Date
import Control.Monad.Except (except)
import Data.Bifunctor (lmap)
import Data.Date as Date
import Data.DateTime as DT
import Data.Formatter.DateTime (FormatterCommand(..), format, unformatDateTime)
import Data.Int (floor)
import Data.List as L
import Data.List.NonEmpty as LNE
import Data.Maybe (fromMaybe)
import Data.Newtype (class Newtype, over, unwrap)
import Data.Time.Duration (Days(..), Milliseconds)
import Effect (Effect)
import Effect.Now as Effect
import Foreign (F, ForeignError(..))
import Prelude (class Eq, class Semigroup, class Show, bind, bottom, const, identity, negate, pure, ($), (/), (<$>), (>>=), (>>>))
import Simple.JSON (class ReadForeign, class WriteForeign)
import Simple.JSON as F

newtype DateRange = DateRange
  { from :: Date
  , to   :: Date
  }


lastNDays :: Days -> Effect DateRange
lastNDays days = do
  now <- Effect.nowDate
  pure $ adjustFrom days $ DateRange { from: now, to: now }


dayCount :: DateRange -> Int
dayCount (DateRange { from, to }) =
  let (days :: Milliseconds) = Date.diff to from
  in floor $ (unwrap days) / 86400000.0


slide :: Days -> DateRange -> DateRange
slide days d = mapDates (adjustDate days >>> sub1) (const $ sub1 (unwrap d).from) d
  where sub1 = adjustDate (Days $ -1.0)


adjustFrom :: Days -> DateRange -> DateRange
adjustFrom days = mapDates (adjustDate days) identity


adjustTo :: Days -> DateRange -> DateRange
adjustTo days = mapDates identity (adjustDate days)


adjustDate :: Days -> Date -> Date
adjustDate days d = fromMaybe d $ Date.adjust days d


mapDates :: (Date -> Date) -> (Date -> Date) -> DateRange -> DateRange
mapDates updFrom updTo = over DateRange (\rec -> rec { from = updFrom rec.from
                                                     , to = updTo rec.to
                                                     })



derive newtype instance showDateRange :: Show DateRange
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
  append (DateRange { from }) (DateRange { to }) =
    DateRange { from, to }
