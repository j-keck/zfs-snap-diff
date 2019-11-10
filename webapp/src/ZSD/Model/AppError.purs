module ZSD.Model.AppError where

import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Prelude (class Show)


data AppError =
    HTTPError HTTPErrors
  | GenericError String
  | Bug String

derive instance genericAppError :: Generic AppError _
instance showAppError :: Show AppError where
  show = genericShow


data HTTPErrors =
    BadRequest String
  | Unauthorized
  | Forbidden
  | NotFound
  | ServerError String
  | ResponseFormatError String
  | RequestError String
  | JSONError String

derive instance genericBackendErrors :: Generic HTTPErrors _
instance showBackendErrors :: Show HTTPErrors where
  show = genericShow
