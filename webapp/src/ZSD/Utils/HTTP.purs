-- | Simple HTTP module - supports `get` and `post` requests.
module ZSD.Utils.HTTP
       ( get
       , get'
       , post
       , post'
       , post_
       ) where

import Prelude

import Affjax as A
import Affjax.RequestBody as ARB
import Affjax.ResponseFormat (ResponseFormat)
import Affjax.ResponseFormat as ARF
import Affjax.StatusCode (StatusCode(..))
import Data.Bifunctor (lmap)
import Data.Either (Either(..))
import Data.List.NonEmpty as LNE
import Data.Maybe (Maybe(..))
import Effect.Aff (Aff)
import Foreign as F
import Simple.JSON (class ReadForeign, class WriteForeign, readJSON, writeJSON)
import Unsafe.Coerce (unsafeCoerce)

import ZSD.Model.AppError (AppError(..), HTTPErrors(..))


type URL = String


-- | performes a get request and returns the response as the given `Affjax.ResponseFormat`
get :: forall a. ResponseFormat a -> URL -> Aff (Either AppError a)
get rfmt url = interpret url <$> A.get rfmt url


-- | performes a get request and decodes the json response
get' :: forall a. ReadForeign a => URL -> Aff (Either AppError a)
get' url = (_ >>= decode url) <$> get ARF.string url


-- | performes a post request with the given payload and returns the response as the given `Affjax.ResponseFormat`
post :: forall a b. WriteForeign a => ResponseFormat b -> URL -> a -> Aff (Either AppError b)
post rfmt url payload = interpret url <$> A.post rfmt url (Just <<< ARB.string <<< writeJSON $ payload)


-- | performes a post request with the given payload and decodes the json response
post' :: forall a b. WriteForeign a => ReadForeign b => URL -> a -> Aff (Either AppError b)
post' url payload = (_ >>= decode url) <$> post ARF.string url payload

-- | performs a post request and ignores the response
post_ :: forall a. WriteForeign a => URL -> a -> Aff (Either AppError Unit)
post_ url payload = post ARF.ignore url payload

-- | decodes the given json string as an instance
decode :: forall a. ReadForeign a => URL -> String -> Either AppError a
decode url = lmap (LNE.head >>> F.renderForeignError >>> JSONError >>> HTTPError url) <<< readJSON


-- | interprets the server response
-- FIXME: error message
interpret :: forall a. URL -> Either A.Error (A.Response a) -> Either AppError a
interpret url = case _ of
  Left err -> Left <<< HTTPError url <<< RequestError $ A.printError err
  Right r -> handleResponse r.status r.body
    where handleResponse (StatusCode code) body
            | code >= 200 && code < 300 = Right body
            | code == 400              = backendError $ BadRequest (unsafeCoerce body)
            | code == 401              = backendError $ Unauthorized
            | code == 403              = backendError $ Forbidden
            | code == 404              = backendError $ NotFound
            | otherwise                = backendError $ ServerError (unsafeCoerce body)
          backendError = Left <<< (HTTPError url)
