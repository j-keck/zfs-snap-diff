-- | Formatter utilities
module ZSD.Utils.Formatter where

import Data.Array as A
import Data.Date (Date)
import Data.DateTime as DT
import Data.Either (fromRight)
import Data.Formatter.DateTime (format, parseFormatString)
import Data.Maybe (Maybe(..))
import Data.Newtype (unwrap)
import Data.Number.Format (fixed, toStringWith)
import Partial.Unsafe (unsafePartial)
import Prelude (bottom, otherwise, ($), (/), (<), (<<<), (<>), (>))

import ZSD.Model.DateTime (DateTime)


-- | formats the given `Number` as a filesize with a unit prefix
filesize :: Number -> String
filesize = go ["B", "K" ,"M", "G", "T", "P"]
  where go us n = case A.uncons us of
          Just { head, tail } -> if(n / 1024.0) > 0.9 then
                                   go tail $ n / 1024.0
                                 else
                                   toS n <> head
          Nothing -> toS n <> "E"



-- | formats the given `ZSD.Model.DateTime` as "dd MMM DD HH:mm YYYY"
dateTime :: DateTime -> String
dateTime = format fmt <<< unwrap
  where fmt = unsafePartial $ fromRight <<< parseFormatString $ "ddd MMM DD HH:mm YYYY"


date :: Date -> String
date d = format fmt $ DT.DateTime d bottom
  where fmt = unsafePartial $ fromRight <<< parseFormatString $ "DD MMM YYYY"


-- | formats a duration in nanos
duration :: Number -> String
duration n
  | n < 1000000000.0    = toS (n / 1000000.0) <> "ms"
  | n < 1000000000000.0 = toS (n / 1000000000.0) <> "s"
  | otherwise           = toS (n / 1000000000000.0) <> "m"


toS :: Number -> String
toS = toStringWith (fixed 1)
