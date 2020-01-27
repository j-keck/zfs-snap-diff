module ZSD.Model.AppError where

import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Prelude (class Show, show, (<>))


data AppError =
    HTTPError HTTPErrors
  | GenericError String
  | Bug String

instance showAppError :: Show AppError where
  show = case _ of
    HTTPError err -> "HTTP error: " <> show err
    GenericError msg -> msg
    Bug msg -> "Unexpected interal state: " <> msg


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
