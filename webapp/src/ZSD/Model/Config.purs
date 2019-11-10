module ZSD.Model.Config where

import ZSD.Model.Dataset

import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)

type Config =
  { datasets :: Datasets
  }


-- | fetches the config from the server
fetch :: Aff (Either AppError Config)
fetch = HTTP.get' "/api/config"
