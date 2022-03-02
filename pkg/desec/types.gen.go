// Package Desec provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.8.3 DO NOT EDIT.
package desec

import (
	"time"

	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
)

// Defines values for CaptchaKind.
const (
	CaptchaKindAudio CaptchaKind = "audio"

	CaptchaKindImage CaptchaKind = "image"
)

// Defines values for DonationInterval.
const (
	DonationIntervalN0 DonationInterval = 0

	DonationIntervalN1 DonationInterval = 1

	DonationIntervalN3 DonationInterval = 3
)

// AuthenticatedActivateUserAction defines model for AuthenticatedActivateUserAction.
type AuthenticatedActivateUserAction struct {
	Captcha *struct {
		Id       string `json:"id"`
		Solution string `json:"solution"`
	} `json:"captcha,omitempty"`
	Domain *string `json:"domain"`
	State  string  `json:"state"`
	User   string  `json:"user"`
}

// AuthenticatedChangeEmailUserAction defines model for AuthenticatedChangeEmailUserAction.
type AuthenticatedChangeEmailUserAction struct {
	NewEmail openapi_types.Email `json:"new_email"`
	State    string              `json:"state"`
	User     string              `json:"user"`
}

// AuthenticatedDeleteUserAction defines model for AuthenticatedDeleteUserAction.
type AuthenticatedDeleteUserAction struct {
	State string `json:"state"`
	User  string `json:"user"`
}

// AuthenticatedRenewDomainBasicUserAction defines model for AuthenticatedRenewDomainBasicUserAction.
type AuthenticatedRenewDomainBasicUserAction struct {
	Domain int    `json:"domain"`
	State  string `json:"state"`
	User   string `json:"user"`
}

// AuthenticatedResetPasswordUserAction defines model for AuthenticatedResetPasswordUserAction.
type AuthenticatedResetPasswordUserAction struct {
	NewPassword string `json:"new_password"`
	State       string `json:"state"`
	User        string `json:"user"`
}

// Captcha defines model for Captcha.
type Captcha struct {
	Challenge *string      `json:"challenge,omitempty"`
	Content   *string      `json:"content,omitempty"`
	Id        *string      `json:"id,omitempty"`
	Kind      *CaptchaKind `json:"kind,omitempty"`
}

// CaptchaKind defines model for Captcha.Kind.
type CaptchaKind string

// ChangeEmail defines model for ChangeEmail.
type ChangeEmail struct {
	NewEmail openapi_types.Email `json:"new_email"`
}

// Domain defines model for Domain.
type Domain struct {
	Created    *time.Time `json:"created,omitempty"`
	Keys       *[]Key     `json:"keys,omitempty"`
	MinimumTtl *int       `json:"minimum_ttl,omitempty"`
	Name       string     `json:"name"`
	Published  *time.Time `json:"published,omitempty"`
	Touched    *string    `json:"touched,omitempty"`
}

// Donation defines model for Donation.
type Donation struct {
	Amount   string               `json:"amount"`
	Bic      *string              `json:"bic,omitempty"`
	Email    *openapi_types.Email `json:"email,omitempty"`
	Iban     string               `json:"iban"`
	Interval *DonationInterval    `json:"interval,omitempty"`
	Message  *string              `json:"message,omitempty"`
	Mref     *string              `json:"mref,omitempty"`
	Name     string               `json:"name"`
}

// DonationInterval defines model for Donation.Interval.
type DonationInterval int

// Key defines model for Key.
type Key struct {
	Dnskey  *string   `json:"dnskey,omitempty"`
	Ds      *[]string `json:"ds,omitempty"`
	Flags   *int      `json:"flags,omitempty"`
	Keytype *string   `json:"keytype,omitempty"`
}

// RRset defines model for RRset.
type RRset struct {
	Created *time.Time `json:"created,omitempty"`
	Domain  *string    `json:"domain,omitempty"`
	Name    *string    `json:"name,omitempty"`
	Records []string   `json:"records"`
	Subname *string    `json:"subname,omitempty"`
	Touched *time.Time `json:"touched,omitempty"`
	Ttl     int        `json:"ttl"`
	Type    string     `json:"type"`
}

// RegisterAccount defines model for RegisterAccount.
type RegisterAccount struct {
	Captcha *struct {
		Id       string `json:"id"`
		Solution string `json:"solution"`
	} `json:"captcha,omitempty"`
	Domain   *string             `json:"domain,omitempty"`
	Email    openapi_types.Email `json:"email"`
	Password *string             `json:"password"`
}

// ResetPassword defines model for ResetPassword.
type ResetPassword struct {
	Captcha struct {
		Id       string `json:"id"`
		Solution string `json:"solution"`
	} `json:"captcha"`
	Email openapi_types.Email `json:"email"`
}

// Token defines model for Token.
type Token struct {
	AllowedSubnets   *[]string  `json:"allowed_subnets,omitempty"`
	Created          *time.Time `json:"created,omitempty"`
	Id               *string    `json:"id,omitempty"`
	IsValid          *string    `json:"is_valid,omitempty"`
	LastUsed         *time.Time `json:"last_used,omitempty"`
	MaxAge           *string    `json:"max_age"`
	MaxUnusedPeriod  *string    `json:"max_unused_period"`
	Name             *string    `json:"name,omitempty"`
	PermManageTokens *bool      `json:"perm_manage_tokens,omitempty"`
}

// User defines model for User.
type User struct {
	Created      *time.Time          `json:"created,omitempty"`
	Email        openapi_types.Email `json:"email"`
	Id           *string             `json:"id,omitempty"`
	LimitDomains *int                `json:"limit_domains"`
	Password     *string             `json:"password"`
}

// CreateRegisterAccountJSONBody defines parameters for CreateRegisterAccount.
type CreateRegisterAccountJSONBody RegisterAccount

// CreateChangeEmailJSONBody defines parameters for CreateChangeEmail.
type CreateChangeEmailJSONBody ChangeEmail

// CreateAccountDeleteJSONBody defines parameters for CreateAccountDelete.
type CreateAccountDeleteJSONBody interface{}

// CreateResetPasswordJSONBody defines parameters for CreateResetPassword.
type CreateResetPasswordJSONBody ResetPassword

// CreateTokenFromLoginJSONBody defines parameters for CreateTokenFromLogin.
type CreateTokenFromLoginJSONBody Token

// CreateAccountLogoutJSONBody defines parameters for CreateAccountLogout.
type CreateAccountLogoutJSONBody interface{}

// ListTokensParams defines parameters for ListTokens.
type ListTokensParams struct {
	// The pagination cursor value.
	Cursor *int `json:"cursor,omitempty"`
}

// CreateTokenJSONBody defines parameters for CreateToken.
type CreateTokenJSONBody Token

// PartialUpdateTokenJSONBody defines parameters for PartialUpdateToken.
type PartialUpdateTokenJSONBody Token

// UpdateTokenJSONBody defines parameters for UpdateToken.
type UpdateTokenJSONBody Token

// CreateCaptchaJSONBody defines parameters for CreateCaptcha.
type CreateCaptchaJSONBody Captcha

// ListDomainsParams defines parameters for ListDomains.
type ListDomainsParams struct {
	// The pagination cursor value.
	Cursor *int `json:"cursor,omitempty"`
}

// CreateDomainJSONBody defines parameters for CreateDomain.
type CreateDomainJSONBody Domain

// ListRRsetsParams defines parameters for ListRRsets.
type ListRRsetsParams struct {
	// The subname filter.
	Subname *string `json:"subname,omitempty"`

	// The record type filter.
	Type *string `json:"type,omitempty"`

	// The pagination cursor value.
	Cursor *int `json:"cursor,omitempty"`
}

// PartialUpdateRRsetsJSONBody defines parameters for PartialUpdateRRsets.
type PartialUpdateRRsetsJSONBody RRset

// CreateRRsetsJSONBody defines parameters for CreateRRsets.
type CreateRRsetsJSONBody RRset

// UpdateRRsetsJSONBody defines parameters for UpdateRRsets.
type UpdateRRsetsJSONBody RRset

// PartialUpdateRRsetJSONBody defines parameters for PartialUpdateRRset.
type PartialUpdateRRsetJSONBody RRset

// UpdateRRsetJSONBody defines parameters for UpdateRRset.
type UpdateRRsetJSONBody RRset

// CreateDonationJSONBody defines parameters for CreateDonation.
type CreateDonationJSONBody Donation

// ListDyndnsRRsetsParams defines parameters for ListDyndnsRRsets.
type ListDyndnsRRsetsParams struct {
	// The pagination cursor value.
	Cursor *int `json:"cursor,omitempty"`
}

// CreateAuthenticatedActivateUserActionJSONBody defines parameters for CreateAuthenticatedActivateUserAction.
type CreateAuthenticatedActivateUserActionJSONBody AuthenticatedActivateUserAction

// CreateAuthenticatedChangeEmailUserActionJSONBody defines parameters for CreateAuthenticatedChangeEmailUserAction.
type CreateAuthenticatedChangeEmailUserActionJSONBody AuthenticatedChangeEmailUserAction

// CreateAuthenticatedDeleteUserActionJSONBody defines parameters for CreateAuthenticatedDeleteUserAction.
type CreateAuthenticatedDeleteUserActionJSONBody AuthenticatedDeleteUserAction

// CreateAuthenticatedRenewDomainBasicUserActionJSONBody defines parameters for CreateAuthenticatedRenewDomainBasicUserAction.
type CreateAuthenticatedRenewDomainBasicUserActionJSONBody AuthenticatedRenewDomainBasicUserAction

// CreateAuthenticatedResetPasswordUserActionJSONBody defines parameters for CreateAuthenticatedResetPasswordUserAction.
type CreateAuthenticatedResetPasswordUserActionJSONBody AuthenticatedResetPasswordUserAction

// CreateRegisterAccountJSONRequestBody defines body for CreateRegisterAccount for application/json ContentType.
type CreateRegisterAccountJSONRequestBody CreateRegisterAccountJSONBody

// CreateChangeEmailJSONRequestBody defines body for CreateChangeEmail for application/json ContentType.
type CreateChangeEmailJSONRequestBody CreateChangeEmailJSONBody

// CreateAccountDeleteJSONRequestBody defines body for CreateAccountDelete for application/json ContentType.
type CreateAccountDeleteJSONRequestBody CreateAccountDeleteJSONBody

// CreateResetPasswordJSONRequestBody defines body for CreateResetPassword for application/json ContentType.
type CreateResetPasswordJSONRequestBody CreateResetPasswordJSONBody

// CreateTokenFromLoginJSONRequestBody defines body for CreateTokenFromLogin for application/json ContentType.
type CreateTokenFromLoginJSONRequestBody CreateTokenFromLoginJSONBody

// CreateAccountLogoutJSONRequestBody defines body for CreateAccountLogout for application/json ContentType.
type CreateAccountLogoutJSONRequestBody CreateAccountLogoutJSONBody

// CreateTokenJSONRequestBody defines body for CreateToken for application/json ContentType.
type CreateTokenJSONRequestBody CreateTokenJSONBody

// PartialUpdateTokenJSONRequestBody defines body for PartialUpdateToken for application/json ContentType.
type PartialUpdateTokenJSONRequestBody PartialUpdateTokenJSONBody

// UpdateTokenJSONRequestBody defines body for UpdateToken for application/json ContentType.
type UpdateTokenJSONRequestBody UpdateTokenJSONBody

// CreateCaptchaJSONRequestBody defines body for CreateCaptcha for application/json ContentType.
type CreateCaptchaJSONRequestBody CreateCaptchaJSONBody

// CreateDomainJSONRequestBody defines body for CreateDomain for application/json ContentType.
type CreateDomainJSONRequestBody CreateDomainJSONBody

// PartialUpdateRRsetsJSONRequestBody defines body for PartialUpdateRRsets for application/json ContentType.
type PartialUpdateRRsetsJSONRequestBody PartialUpdateRRsetsJSONBody

// CreateRRsetsJSONRequestBody defines body for CreateRRsets for application/json ContentType.
type CreateRRsetsJSONRequestBody CreateRRsetsJSONBody

// UpdateRRsetsJSONRequestBody defines body for UpdateRRsets for application/json ContentType.
type UpdateRRsetsJSONRequestBody UpdateRRsetsJSONBody

// PartialUpdateRRsetJSONRequestBody defines body for PartialUpdateRRset for application/json ContentType.
type PartialUpdateRRsetJSONRequestBody PartialUpdateRRsetJSONBody

// UpdateRRsetJSONRequestBody defines body for UpdateRRset for application/json ContentType.
type UpdateRRsetJSONRequestBody UpdateRRsetJSONBody

// CreateDonationJSONRequestBody defines body for CreateDonation for application/json ContentType.
type CreateDonationJSONRequestBody CreateDonationJSONBody

// CreateAuthenticatedActivateUserActionJSONRequestBody defines body for CreateAuthenticatedActivateUserAction for application/json ContentType.
type CreateAuthenticatedActivateUserActionJSONRequestBody CreateAuthenticatedActivateUserActionJSONBody

// CreateAuthenticatedChangeEmailUserActionJSONRequestBody defines body for CreateAuthenticatedChangeEmailUserAction for application/json ContentType.
type CreateAuthenticatedChangeEmailUserActionJSONRequestBody CreateAuthenticatedChangeEmailUserActionJSONBody

// CreateAuthenticatedDeleteUserActionJSONRequestBody defines body for CreateAuthenticatedDeleteUserAction for application/json ContentType.
type CreateAuthenticatedDeleteUserActionJSONRequestBody CreateAuthenticatedDeleteUserActionJSONBody

// CreateAuthenticatedRenewDomainBasicUserActionJSONRequestBody defines body for CreateAuthenticatedRenewDomainBasicUserAction for application/json ContentType.
type CreateAuthenticatedRenewDomainBasicUserActionJSONRequestBody CreateAuthenticatedRenewDomainBasicUserActionJSONBody

// CreateAuthenticatedResetPasswordUserActionJSONRequestBody defines body for CreateAuthenticatedResetPasswordUserAction for application/json ContentType.
type CreateAuthenticatedResetPasswordUserActionJSONRequestBody CreateAuthenticatedResetPasswordUserActionJSONBody
