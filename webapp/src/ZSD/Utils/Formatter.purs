-- | Formatter utilities
module ZSD.Utils.Formatter where

import Data.Array as A
import Data.Date (Date)
import Data.DateTime as DT
import Data.Either (either)
import Data.Formatter.DateTime (formatDateTime)
import Data.Maybe (Maybe(..))
import Data.Newtype (unwrap)
import Data.Number.Format (fixed, toStringWith)
import Partial.Unsafe (unsafeCrashWith)
import Prelude (bottom, identity, otherwise, ($), (/), (<), (<<<), (<>), (>))
import ZSD.Model.DateTime (DateTime)

-- | formats the given `Number` as a filesize with a unit prefix
filesize :: Number -> String
filesize = go [ "B", "K", "M", "G", "T", "P" ]
  where
  go us n = case A.uncons us of
    Just { head, tail } ->
      if (n / 1024.0) > 0.9 then
        go tail $ n / 1024.0
      else
        toS n <> head
    Nothing -> toS n <> "E"

-- | formats the given `ZSD.Model.DateTime` as "dd MMM DD HH:mm YYYY"
dateTime :: DateTime -> String
dateTime = fmt <<< unwrap
  where
  fmt dt = either (\msg -> unsafeCrashWith "Invalid dateTime: " <> msg) identity $ formatDateTime "ddd MMM DD HH:mm YYYY" dt

date :: Date -> String
date d = fmt $ DT.DateTime d bottom
  where
  fmt dt = either (\msg -> unsafeCrashWith "Invalid date: " <> msg) identity $ formatDateTime "DD MMM YYYY" dt

-- | formats a duration in nanos
duration :: Number -> String
duration n
  | n < 1000000000.0 = toS (n / 1000000.0) <> "ms"
  | n < 1000000000000.0 = toS (n / 1000000000.0) <> "s"
  | otherwise = toS (n / 1000000000000.0) <> "m"

toS :: Number -> String
toS = toStringWith (fixed 1)
