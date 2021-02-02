package models

import (
	"database/sql"
	"strings"

	"github.com/gobuffalo/pop/nulls"
	"github.com/gobuffalo/uuid"
	"github.com/netlify/gotrue/crypto"
	"github.com/netlify/gotrue/storage"
	"github.com/netlify/gotrue/storage/namespace"
	"github.com/pkg/errors"
)

type Identity struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	UserId    uuid.UUID    `json:"userID" db:"user_id"`
	AccessKey string       `json:"accessKey" db:"access_key"`
	SecretKey string       `json:"secretKey" db:"secret_key"`
	Token     nulls.String `json:"token" db:"user_token"`
}

func NewIdentity(userId uuid.UUID) (*Identity, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating unique id")
	}
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating unique id")
	}
	key := strings.Replace(uid.String(), "-", "", 4)

	identity := &Identity{
		ID:        id,
		UserId:    userId,
		AccessKey: crypto.SecureToken(),
		SecretKey: key,
	}

	return identity, nil
}

func (Identity) TableName() string {
	tableName := "identities"

	if namespace.GetNamespace() != "" {
		return namespace.GetNamespace() + "." + tableName
	}

	return tableName
}

func (i *Identity) UpdateIdentityToken(tx *storage.Connection, token string) error {
	i.Token = nulls.NewString(token)

	return tx.UpdateOnly(i, "user_token")
}

func (i *Identity) UpdateIdentityAccessKey(tx *storage.Connection) error {
	i.AccessKey = crypto.SecureToken()

	return tx.UpdateOnly(i, "access_key")
}

func (i *Identity) UpdateIdentitySecret(tx *storage.Connection) error {
	uid, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "Error generating unique id")
	}
	key := strings.Replace(uid.String(), "-", "", 4)

	i.SecretKey = key

	return tx.UpdateOnly(i, "secret_key")
}

func findIdentity(tx *storage.Connection, query string, args ...interface{}) (*Identity, error) {
	obj := &Identity{}

	if err := tx.Q().Where(query, args...).First(obj); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, IdentityNotFoundError{}
		}
		return nil, errors.Wrap(err, "error finding identity")
	}

	return obj, nil
}

func FindIdentityByID(tx *storage.Connection, id uuid.UUID) (*Identity, error) {
	return findIdentity(tx, "id = ?", id)
}

func FindIdentityByUser(tx *storage.Connection, id uuid.UUID) (*Identity, error) {
	return findIdentity(tx, "user_id = ?", id)
}

func FindIdentityByToken(tx *storage.Connection, token string) (*Identity, error) {
	return findIdentity(tx, "user_token = ?", token)
}
