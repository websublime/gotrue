package api

import (
	"encoding/json"
	"net/http"

	"github.com/netlify/gotrue/models"
	"github.com/netlify/gotrue/storage"
)

type TemplateParams struct {
	Type    models.TemplateType `json:"type"`
	Subject string              `json:"subject"`
	Url     string              `json:"url"`
}

func (a *API) CreateTemplate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	claims := getClaims(ctx)
	if claims == nil {
		return badRequestError("Could not read claims")
	}

	params := &TemplateParams{}

	aud := a.requestAud(ctx, r)
	if aud != claims.Audience {
		return badRequestError("Token audience doesn't match request audience")
	}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)
	if err != nil {
		return badRequestError("Could not read Template params: %v", err)
	}

	types := []string{
		string(models.Invite),
		string(models.Confirmation),
		string(models.Email),
		string(models.Recovery),
	}

	ok := contains(types, string(params.Type))

	if !ok {
		return badRequestError("Template type not found")
	}

	template, err := models.NewTemplate(aud, params.Type, params.Subject, params.Url)
	if err != nil {
		return internalServerError("Database error creating template").WithInternalError(err)
	}

	a.db.Transaction(func(tx *storage.Connection) error {
		if terr := tx.Create(template); terr != nil {
			return internalServerError("Database error saving new template").WithInternalError(terr)
		}

		return nil
	})

	return sendJSON(w, http.StatusOK, template)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
