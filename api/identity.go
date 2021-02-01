package api

import (
	"net/http"

	"github.com/gobuffalo/uuid"
	"github.com/netlify/gotrue/models"
)

func (a *API) IdentityGet(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	claims := getClaims(ctx)

	if claims == nil {
		return badRequestError("Could not read claims")
	}

	userID, err := uuid.FromString(claims.Subject)
	if err != nil {
		return badRequestError("Could not read User ID claim")
	}

	aud := a.requestAud(ctx, r)
	if aud != claims.Audience {
		return badRequestError("Token audience doesn't match request audience")
	}

	user, err := models.FindUserByID(a.db, userID)
	if err != nil {
		if models.IsNotFoundError(err) {
			return notFoundError(err.Error())
		}
		return internalServerError("Database error finding user").WithInternalError(err)
	}

	identity, err := models.FindIdentityByUser(a.db, user.ID)
	if err != nil {
		if models.IsNotFoundError(err) {
			return notFoundError(err.Error())
		}
		return internalServerError("Database error finding identity").WithInternalError(err)
	}

	return sendJSON(w, http.StatusOK, identity)

}
