module ZSD.Model.Config where

import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.Model.AppError (AppError)
import ZSD.Utils.HTTP as HTTP

type Config
  = { daysToScan :: Int
    , snapshotNameTemplate :: String
    }


-- | fetches the config from the server
fetch :: Aff (Either AppError Config)
fetch = HTTP.get' "api/config"
