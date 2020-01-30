module ZSD.Model.AppError where

import Affjax (URL)
import Prelude (class Show, show, (<>))


data AppError =
    HTTPError URL HTTPErrors
  | GenericError String
  | Bug String

instance showAppError :: Show AppError where
  show = case _ of
    HTTPError url err -> "client <-> server communication error at endpoint: '"
                         <> url <> "' - "
                         <> show err
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

instance showHTTPErrors :: Show HTTPErrors where
  show = case _ of
    BadRequest msg -> "bad request: " <> msg
    Unauthorized -> "unauthorized"
    Forbidden -> "forbidden"
    NotFound -> "resource not found"
    ServerError msg -> "server error: " <> msg
    ResponseFormatError msg -> "response format error: " <> msg
    RequestError msg -> "request error: " <> msg
    JSONError msg -> "json error: " <> msg
