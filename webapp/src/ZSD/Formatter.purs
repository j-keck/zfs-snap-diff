-- | Formatter utilities
module ZSD.Formatter where

import Data.Array as A
import Data.Either (fromRight)
import Data.Formatter.DateTime (format, parseFormatString)
import Data.Maybe (Maybe(..))
import Data.Newtype (unwrap)
import Data.Number.Format (fixed, precision, toStringWith)
import Partial.Unsafe (unsafePartial)
import Prelude (($), (/), (<<<), (<>), (>))
import ZSD.Model.DateTime (DateTime)


-- | formats the given `Number` as a filesize with a unit prefix
filesize :: Number -> String
filesize = go ["B", "K" ,"M", "G", "T", "P"]
  where go us n = case A.uncons us of
          Just { head, tail } -> if(n / 1024.0) > 0.9 then
                                   go tail $ n / 1024.0
                                 else
                                   toStringWith (fixed 1) n <> head
          Nothing -> toStringWith (precision 1)  n <> "E"



-- | formats the given `ZSD.Model.DateTime` as "ddd MMM DD HH:mm YYYY"
dateTime :: DateTime -> String
dateTime = format fmt <<< unwrap
  where fmt = unsafePartial $ fromRight <<< parseFormatString $ "ddd MMM DD HH:mm YYYY"
