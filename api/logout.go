package api

import (
	"context"
	"net/http"

	"github.com/netlify/gotrue/models"
	"github.com/netlify/gotrue/storage"
)

// Logout is the endpoint for logging out a user and thereby revoking any refresh tokens
func (a *API) Logout(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	instanceID := getInstanceID(ctx)

	a.clearCookieToken(ctx, w)

	u, err := getUserFromClaims(ctx, a.db)
	if err != nil {
		return unauthorizedError("Invalid user").WithInternalError(err)
	}

	err = a.db.Transaction(func(tx *storage.Connection) error {
		if terr := models.NewAuditLogEntry(tx, instanceID, u, models.LogoutAction, nil); terr != nil {
			return terr
		}
		return models.Logout(tx, instanceID, u.ID)
	})
	if err != nil {
		return internalServerError("Error logging out user").WithInternalError(err)
	}

	err = a.identityResetTotken(ctx, u)
	if err != nil {
		return internalServerError("Error clear out user identity").WithInternalError(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (a *API) identityResetTotken(ctx context.Context, user *models.User) error {
	var identity *models.Identity

	err := a.db.Transaction(func(tx *storage.Connection) error {
		var terr error
		identity, terr = models.FindIdentityByUser(tx, user.ID)
		if terr != nil {
			return internalServerError("Database error identity user").WithInternalError(terr)
		}

		terr = identity.UpdateIdentityToken(tx, "")
		if terr != nil {
			return internalServerError("Databse error reseting identity user").WithInternalError(terr)
		}

		return terr
	})

	return err
}
