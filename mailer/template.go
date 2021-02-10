package mailer

import (
	"github.com/badoux/checkmail"
	"github.com/netlify/gotrue/conf"
	"github.com/netlify/gotrue/models"
	"github.com/netlify/mailme"
)

// TemplateMailer will send mail and use templates from the site for easy mail styling
type TemplateMailer struct {
	SiteURL string
	Config  *conf.Configuration
	Mailer  *mailme.Mailer
}

type TemplateMailData struct {
	TemplateURL string
	CallbackUrl string
	Subject     string
}

const defaultInviteMail = `<h2>You have been invited</h2>

<p>You have been invited to create a user on {{ .SiteURL }}. Follow this link to accept the invite:</p>
<p><a href="{{ .ConfirmationURL }}">Accept the invite</a></p>`

const defaultConfirmationMail = `<h2>Confirm your signup</h2>

<p>Follow this link to confirm your user:</p>
<p><a href="{{ .ConfirmationURL }}">Confirm your email address</a></p>`

const defaultRecoveryMail = `<h2>Reset password</h2>

<p>Follow this link to reset the password for your user:</p>
<p><a href="{{ .ConfirmationURL }}">Reset password</a></p>`

const defaultEmailChangeMail = `<h2>Confirm email address change</h2>

<p>Follow this link to confirm the update of your email address from {{ .Email }} to {{ .NewEmail }}:</p>
<p><a href="{{ .ConfirmationURL }}">Change email address</a></p>`

// ValidateEmail returns nil if the email is valid,
// otherwise an error indicating the reason it is invalid
func (m TemplateMailer) ValidateEmail(email string) error {
	return checkmail.ValidateFormat(email)
}

func (m TemplateMailer) TemplateMail(template *models.Template, subjectDefault string, urlTplDefault string) *TemplateMailData {
	if len(template.BaseURL) > 0 {
		m.Mailer.BaseURL = template.BaseURL
		m.Config.SiteURL = template.BaseURL
	}

	tpl := &TemplateMailData{}

	if len(template.Url) > 0 {
		tpl.CallbackUrl = template.Url
	} else {
		tpl.CallbackUrl = "/"
	}

	if len(template.Subject) > 0 {
		tpl.Subject = template.Subject
	} else {
		tpl.Subject = subjectDefault
	}

	if len(template.UrlTemplate) > 0 {
		tpl.TemplateURL = template.UrlTemplate
	} else {
		tpl.TemplateURL = urlTplDefault
	}

	return tpl
}

// InviteMail sends a invite mail to a new user
func (m *TemplateMailer) InviteMail(user *models.User, template *models.Template, referrerURL string) error {
	templateData := m.TemplateMail(template, m.Config.Mailer.Subjects.Invite, m.Config.Mailer.Templates.Invite)

	url, err := getSiteURL(referrerURL, m.Config.SiteURL, templateData.CallbackUrl, "invite_token="+user.ConfirmationToken)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"SiteURL":         m.Config.SiteURL,
		"ConfirmationURL": url,
		"Email":           user.Email,
		"Token":           user.ConfirmationToken,
		"Data":            user.UserMetaData,
	}

	// TODO: test it
	return m.Mailer.Mail(
		user.Email,
		string(withDefault(templateData.Subject, "You have been invited")),
		enforceRelativeURL(templateData.TemplateURL),
		defaultInviteMail,
		data,
	)
}

// ConfirmationMail sends a signup confirmation mail to a new user
func (m *TemplateMailer) ConfirmationMail(user *models.User, template *models.Template, referrerURL string) error {
	templateData := m.TemplateMail(template, m.Config.Mailer.Subjects.Confirmation, m.Config.Mailer.Templates.Confirmation)

	url, err := getSiteURL(referrerURL, m.Config.SiteURL, templateData.CallbackUrl, "confirmation_token="+user.ConfirmationToken)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"SiteURL":         m.Config.SiteURL,
		"ConfirmationURL": url,
		"Email":           user.Email,
		"Token":           user.ConfirmationToken,
		"Data":            user.UserMetaData,
	}

	return m.Mailer.Mail(
		user.Email,
		string(withDefault(templateData.Subject, "Confirm Your Signup")),
		enforceRelativeURL(templateData.TemplateURL),
		defaultConfirmationMail,
		data,
	)
}

// EmailChangeMail sends an email change confirmation mail to a user
func (m *TemplateMailer) EmailChangeMail(user *models.User, template *models.Template, referrerURL string) error {
	templateData := m.TemplateMail(template, m.Config.Mailer.Subjects.EmailChange, m.Config.Mailer.Templates.EmailChange)
	url, err := getSiteURL(referrerURL, m.Config.SiteURL, templateData.CallbackUrl, "email_change_token="+user.EmailChangeToken)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"SiteURL":         m.Config.SiteURL,
		"ConfirmationURL": url,
		"Email":           user.Email,
		"NewEmail":        user.EmailChange,
		"Token":           user.EmailChangeToken,
		"Data":            user.UserMetaData,
	}

	return m.Mailer.Mail(
		user.EmailChange,
		string(withDefault(templateData.Subject, "Confirm Email Change")),
		enforceRelativeURL(templateData.TemplateURL),
		defaultEmailChangeMail,
		data,
	)
}

// RecoveryMail sends a password recovery mail
func (m *TemplateMailer) RecoveryMail(user *models.User, template *models.Template, referrerURL string) error {
	templateData := m.TemplateMail(template, m.Config.Mailer.Subjects.Recovery, m.Config.Mailer.Templates.Recovery)
	url, err := getSiteURL(referrerURL, m.Config.SiteURL, templateData.CallbackUrl, "recovery_token="+user.RecoveryToken)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"SiteURL":         m.Config.SiteURL,
		"ConfirmationURL": url,
		"Email":           user.Email,
		"Token":           user.RecoveryToken,
		"Data":            user.UserMetaData,
	}

	return m.Mailer.Mail(
		user.Email,
		string(withDefault(templateData.Subject, "Reset Your Password")),
		enforceRelativeURL(templateData.TemplateURL),
		defaultRecoveryMail,
		data,
	)
}

// Send can be used to send one-off emails to users
func (m TemplateMailer) Send(user *models.User, subject, body string, data map[string]interface{}) error {
	return m.Mailer.Mail(
		user.Email,
		subject,
		"",
		body,
		data,
	)
}
