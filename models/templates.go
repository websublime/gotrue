package models

import (
	"database/sql"

	"github.com/gobuffalo/uuid"
	"github.com/netlify/gotrue/storage"
	"github.com/netlify/gotrue/storage/namespace"
	"github.com/pkg/errors"
)

type TemplateType string

const (
	Invite       TemplateType = "INVITE"
	Confirmation TemplateType = "CONFIRMATION"
	Recovery     TemplateType = "RECOVERY"
	Email        TemplateType = "EMAIL"
)

type Template struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Aud         string       `json:"aud" db:"aud"`
	Type        TemplateType `json:"type" db:"type"`
	Subject     string       `json:"subject" db:"subject"`
	Url         string       `json:"url" db:"url"`
	BaseURL     string       `json:"baseUrl" db:"base_url"`
	UrlTemplate string       `json:"urlTemplate" db:"url_template"`
}

func NewTemplate(aud string, types TemplateType, subject string, url string, base string, urlTemplate string) (*Template, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating unique id")
	}

	template := &Template{
		ID:      id,
		Aud:     aud,
		Type:    types,
		Subject: subject,
		Url:     url,
    BaseURL: base,
    UrlTemplate: ,
	}

	return template, nil
}

func (Template) TableName() string {
	tableName := "templates"

	if namespace.GetNamespace() != "" {
		return namespace.GetNamespace() + "." + tableName
	}

	return tableName
}

func FindTemplate(tx *storage.Connection, aud string, types string) (*Template, error) {
	obj := &Template{}
	if err := tx.Q().Where("aud = ?", aud).Where("type = ?", types).First(obj); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, TemplateNotFoundError{}
		}
	}

	return obj, nil
}
