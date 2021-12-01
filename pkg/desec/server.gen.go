// Package Desec provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.8.3 DO NOT EDIT.
package desec

import (
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /api/v1/)
	ListRoots(ctx echo.Context) error

	// (POST /api/v1/auth/)
	CreateRegisterAccount(ctx echo.Context) error

	// (GET /api/v1/auth/account/)
	RetrieveUser(ctx echo.Context) error

	// (POST /api/v1/auth/account/change-email/)
	CreateChangeEmail(ctx echo.Context) error

	// (POST /api/v1/auth/account/delete/)
	CreateAccountDelete(ctx echo.Context) error

	// (POST /api/v1/auth/account/reset-password/)
	CreateResetPassword(ctx echo.Context) error

	// (POST /api/v1/auth/login/)
	CreateTokenFromLogin(ctx echo.Context) error

	// (POST /api/v1/auth/logout/)
	CreateAccountLogout(ctx echo.Context) error

	// (GET /api/v1/auth/tokens/)
	ListTokens(ctx echo.Context, params ListTokensParams) error

	// (POST /api/v1/auth/tokens/)
	CreateToken(ctx echo.Context) error

	// (DELETE /api/v1/auth/tokens/{id}/)
	DestroyToken(ctx echo.Context, id string) error

	// (GET /api/v1/auth/tokens/{id}/)
	RetrieveToken(ctx echo.Context, id string) error

	// (PATCH /api/v1/auth/tokens/{id}/)
	PartialUpdateToken(ctx echo.Context, id string) error

	// (PUT /api/v1/auth/tokens/{id}/)
	UpdateToken(ctx echo.Context, id string) error

	// (POST /api/v1/captcha/)
	CreateCaptcha(ctx echo.Context) error

	// (GET /api/v1/domains/)
	ListDomains(ctx echo.Context, params ListDomainsParams) error

	// (POST /api/v1/domains/)
	CreateDomain(ctx echo.Context) error

	// (DELETE /api/v1/domains/{name}/)
	DestroyDomain(ctx echo.Context, name string) error

	// (GET /api/v1/domains/{name}/)
	RetrieveDomain(ctx echo.Context, name string) error

	// (GET /api/v1/domains/{name}/rrsets/)
	ListRRsets(ctx echo.Context, name string, params ListRRsetsParams) error

	// (PATCH /api/v1/domains/{name}/rrsets/)
	PartialUpdateRRset(ctx echo.Context, name string) error

	// (POST /api/v1/domains/{name}/rrsets/)
	CreateRRset(ctx echo.Context, name string) error

	// (PUT /api/v1/domains/{name}/rrsets/)
	UpdateRRset(ctx echo.Context, name string) error

	// (DELETE /api/v1/domains/{name}/rrsets/@/{type}/)
	DestroyApexRRset(ctx echo.Context, name string, pType string) error

	// (GET /api/v1/domains/{name}/rrsets/@/{type}/)
	RetrieveApexRRset(ctx echo.Context, name string, pType string) error

	// (PATCH /api/v1/domains/{name}/rrsets/@/{type}/)
	PartialUpdateApexRRset(ctx echo.Context, name string, pType string) error

	// (PUT /api/v1/domains/{name}/rrsets/@/{type}/)
	UpdateApexRRset(ctx echo.Context, name string, pType string) error

	// (DELETE /api/v1/domains/{name}/rrsets/{subname}/{type}/)
	DestroySubnameRRset(ctx echo.Context, name string, subname string, pType string) error

	// (GET /api/v1/domains/{name}/rrsets/{subname}/{type}/)
	RetrieveSubnameRRset(ctx echo.Context, name string, subname string, pType string) error

	// (PATCH /api/v1/domains/{name}/rrsets/{subname}/{type}/)
	PartialUpdateSubnameRRset(ctx echo.Context, name string, subname string, pType string) error

	// (PUT /api/v1/domains/{name}/rrsets/{subname}/{type}/)
	UpdateSubnameRRset(ctx echo.Context, name string, subname string, pType string) error

	// (POST /api/v1/donation/)
	CreateDonation(ctx echo.Context) error

	// (GET /api/v1/dyndns/update)
	ListDyndnsRRsets(ctx echo.Context, params ListDyndnsRRsetsParams) error

	// (GET /api/v1/serials/)
	ListSerials(ctx echo.Context) error

	// (GET /api/v1/v/activate-account/{code}/)
	RetrieveAuthenticatedActivateUserAction(ctx echo.Context, code string) error

	// (POST /api/v1/v/activate-account/{code}/)
	CreateAuthenticatedActivateUserAction(ctx echo.Context, code string) error

	// (GET /api/v1/v/change-email/{code}/)
	RetrieveAuthenticatedChangeEmailUserAction(ctx echo.Context, code string) error

	// (POST /api/v1/v/change-email/{code}/)
	CreateAuthenticatedChangeEmailUserAction(ctx echo.Context, code string) error

	// (GET /api/v1/v/delete-account/{code}/)
	RetrieveAuthenticatedDeleteUserAction(ctx echo.Context, code string) error

	// (POST /api/v1/v/delete-account/{code}/)
	CreateAuthenticatedDeleteUserAction(ctx echo.Context, code string) error

	// (GET /api/v1/v/renew-domain/{code}/)
	RetrieveAuthenticatedRenewDomainBasicUserAction(ctx echo.Context, code string) error

	// (POST /api/v1/v/renew-domain/{code}/)
	CreateAuthenticatedRenewDomainBasicUserAction(ctx echo.Context, code string) error

	// (GET /api/v1/v/reset-password/{code}/)
	RetrieveAuthenticatedResetPasswordUserAction(ctx echo.Context, code string) error

	// (POST /api/v1/v/reset-password/{code}/)
	CreateAuthenticatedResetPasswordUserAction(ctx echo.Context, code string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// ListRoots converts echo context to params.
func (w *ServerInterfaceWrapper) ListRoots(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListRoots(ctx)
	return err
}

// CreateRegisterAccount converts echo context to params.
func (w *ServerInterfaceWrapper) CreateRegisterAccount(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateRegisterAccount(ctx)
	return err
}

// RetrieveUser converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveUser(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveUser(ctx)
	return err
}

// CreateChangeEmail converts echo context to params.
func (w *ServerInterfaceWrapper) CreateChangeEmail(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateChangeEmail(ctx)
	return err
}

// CreateAccountDelete converts echo context to params.
func (w *ServerInterfaceWrapper) CreateAccountDelete(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateAccountDelete(ctx)
	return err
}

// CreateResetPassword converts echo context to params.
func (w *ServerInterfaceWrapper) CreateResetPassword(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateResetPassword(ctx)
	return err
}

// CreateTokenFromLogin converts echo context to params.
func (w *ServerInterfaceWrapper) CreateTokenFromLogin(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateTokenFromLogin(ctx)
	return err
}

// CreateAccountLogout converts echo context to params.
func (w *ServerInterfaceWrapper) CreateAccountLogout(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateAccountLogout(ctx)
	return err
}

// ListTokens converts echo context to params.
func (w *ServerInterfaceWrapper) ListTokens(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ListTokensParams
	// ------------- Optional query parameter "cursor" -------------

	err = runtime.BindQueryParameter("form", true, false, "cursor", ctx.QueryParams(), &params.Cursor)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter cursor: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListTokens(ctx, params)
	return err
}

// CreateToken converts echo context to params.
func (w *ServerInterfaceWrapper) CreateToken(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateToken(ctx)
	return err
}

// DestroyToken converts echo context to params.
func (w *ServerInterfaceWrapper) DestroyToken(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DestroyToken(ctx, id)
	return err
}

// RetrieveToken converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveToken(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveToken(ctx, id)
	return err
}

// PartialUpdateToken converts echo context to params.
func (w *ServerInterfaceWrapper) PartialUpdateToken(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PartialUpdateToken(ctx, id)
	return err
}

// UpdateToken converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateToken(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateToken(ctx, id)
	return err
}

// CreateCaptcha converts echo context to params.
func (w *ServerInterfaceWrapper) CreateCaptcha(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateCaptcha(ctx)
	return err
}

// ListDomains converts echo context to params.
func (w *ServerInterfaceWrapper) ListDomains(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ListDomainsParams
	// ------------- Optional query parameter "cursor" -------------

	err = runtime.BindQueryParameter("form", true, false, "cursor", ctx.QueryParams(), &params.Cursor)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter cursor: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListDomains(ctx, params)
	return err
}

// CreateDomain converts echo context to params.
func (w *ServerInterfaceWrapper) CreateDomain(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateDomain(ctx)
	return err
}

// DestroyDomain converts echo context to params.
func (w *ServerInterfaceWrapper) DestroyDomain(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DestroyDomain(ctx, name)
	return err
}

// RetrieveDomain converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveDomain(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveDomain(ctx, name)
	return err
}

// ListRRsets converts echo context to params.
func (w *ServerInterfaceWrapper) ListRRsets(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params ListRRsetsParams
	// ------------- Optional query parameter "subname" -------------

	err = runtime.BindQueryParameter("form", true, false, "subname", ctx.QueryParams(), &params.Subname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter subname: %s", err))
	}

	// ------------- Optional query parameter "type" -------------

	err = runtime.BindQueryParameter("form", true, false, "type", ctx.QueryParams(), &params.Type)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// ------------- Optional query parameter "cursor" -------------

	err = runtime.BindQueryParameter("form", true, false, "cursor", ctx.QueryParams(), &params.Cursor)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter cursor: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListRRsets(ctx, name, params)
	return err
}

// PartialUpdateRRset converts echo context to params.
func (w *ServerInterfaceWrapper) PartialUpdateRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PartialUpdateRRset(ctx, name)
	return err
}

// CreateRRset converts echo context to params.
func (w *ServerInterfaceWrapper) CreateRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateRRset(ctx, name)
	return err
}

// UpdateRRset converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateRRset(ctx, name)
	return err
}

// DestroyApexRRset converts echo context to params.
func (w *ServerInterfaceWrapper) DestroyApexRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DestroyApexRRset(ctx, name, pType)
	return err
}

// RetrieveApexRRset converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveApexRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveApexRRset(ctx, name, pType)
	return err
}

// PartialUpdateApexRRset converts echo context to params.
func (w *ServerInterfaceWrapper) PartialUpdateApexRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PartialUpdateApexRRset(ctx, name, pType)
	return err
}

// UpdateApexRRset converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateApexRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateApexRRset(ctx, name, pType)
	return err
}

// DestroySubnameRRset converts echo context to params.
func (w *ServerInterfaceWrapper) DestroySubnameRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "subname" -------------
	var subname string

	err = runtime.BindStyledParameterWithLocation("simple", false, "subname", runtime.ParamLocationPath, ctx.Param("subname"), &subname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter subname: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DestroySubnameRRset(ctx, name, subname, pType)
	return err
}

// RetrieveSubnameRRset converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveSubnameRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "subname" -------------
	var subname string

	err = runtime.BindStyledParameterWithLocation("simple", false, "subname", runtime.ParamLocationPath, ctx.Param("subname"), &subname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter subname: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveSubnameRRset(ctx, name, subname, pType)
	return err
}

// PartialUpdateSubnameRRset converts echo context to params.
func (w *ServerInterfaceWrapper) PartialUpdateSubnameRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "subname" -------------
	var subname string

	err = runtime.BindStyledParameterWithLocation("simple", false, "subname", runtime.ParamLocationPath, ctx.Param("subname"), &subname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter subname: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PartialUpdateSubnameRRset(ctx, name, subname, pType)
	return err
}

// UpdateSubnameRRset converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateSubnameRRset(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameterWithLocation("simple", false, "name", runtime.ParamLocationPath, ctx.Param("name"), &name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// ------------- Path parameter "subname" -------------
	var subname string

	err = runtime.BindStyledParameterWithLocation("simple", false, "subname", runtime.ParamLocationPath, ctx.Param("subname"), &subname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter subname: %s", err))
	}

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameterWithLocation("simple", false, "type", runtime.ParamLocationPath, ctx.Param("type"), &pType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateSubnameRRset(ctx, name, subname, pType)
	return err
}

// CreateDonation converts echo context to params.
func (w *ServerInterfaceWrapper) CreateDonation(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateDonation(ctx)
	return err
}

// ListDyndnsRRsets converts echo context to params.
func (w *ServerInterfaceWrapper) ListDyndnsRRsets(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ListDyndnsRRsetsParams
	// ------------- Optional query parameter "cursor" -------------

	err = runtime.BindQueryParameter("form", true, false, "cursor", ctx.QueryParams(), &params.Cursor)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter cursor: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListDyndnsRRsets(ctx, params)
	return err
}

// ListSerials converts echo context to params.
func (w *ServerInterfaceWrapper) ListSerials(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListSerials(ctx)
	return err
}

// RetrieveAuthenticatedActivateUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveAuthenticatedActivateUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveAuthenticatedActivateUserAction(ctx, code)
	return err
}

// CreateAuthenticatedActivateUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) CreateAuthenticatedActivateUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateAuthenticatedActivateUserAction(ctx, code)
	return err
}

// RetrieveAuthenticatedChangeEmailUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveAuthenticatedChangeEmailUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveAuthenticatedChangeEmailUserAction(ctx, code)
	return err
}

// CreateAuthenticatedChangeEmailUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) CreateAuthenticatedChangeEmailUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateAuthenticatedChangeEmailUserAction(ctx, code)
	return err
}

// RetrieveAuthenticatedDeleteUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveAuthenticatedDeleteUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveAuthenticatedDeleteUserAction(ctx, code)
	return err
}

// CreateAuthenticatedDeleteUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) CreateAuthenticatedDeleteUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateAuthenticatedDeleteUserAction(ctx, code)
	return err
}

// RetrieveAuthenticatedRenewDomainBasicUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveAuthenticatedRenewDomainBasicUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveAuthenticatedRenewDomainBasicUserAction(ctx, code)
	return err
}

// CreateAuthenticatedRenewDomainBasicUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) CreateAuthenticatedRenewDomainBasicUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateAuthenticatedRenewDomainBasicUserAction(ctx, code)
	return err
}

// RetrieveAuthenticatedResetPasswordUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveAuthenticatedResetPasswordUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveAuthenticatedResetPasswordUserAction(ctx, code)
	return err
}

// CreateAuthenticatedResetPasswordUserAction converts echo context to params.
func (w *ServerInterfaceWrapper) CreateAuthenticatedResetPasswordUserAction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "code" -------------
	var code string

	err = runtime.BindStyledParameterWithLocation("simple", false, "code", runtime.ParamLocationPath, ctx.Param("code"), &code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateAuthenticatedResetPasswordUserAction(ctx, code)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {
	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/api/v1/", wrapper.ListRoots)
	router.POST(baseURL+"/api/v1/auth/", wrapper.CreateRegisterAccount)
	router.GET(baseURL+"/api/v1/auth/account/", wrapper.RetrieveUser)
	router.POST(baseURL+"/api/v1/auth/account/change-email/", wrapper.CreateChangeEmail)
	router.POST(baseURL+"/api/v1/auth/account/delete/", wrapper.CreateAccountDelete)
	router.POST(baseURL+"/api/v1/auth/account/reset-password/", wrapper.CreateResetPassword)
	router.POST(baseURL+"/api/v1/auth/login/", wrapper.CreateTokenFromLogin)
	router.POST(baseURL+"/api/v1/auth/logout/", wrapper.CreateAccountLogout)
	router.GET(baseURL+"/api/v1/auth/tokens/", wrapper.ListTokens)
	router.POST(baseURL+"/api/v1/auth/tokens/", wrapper.CreateToken)
	router.DELETE(baseURL+"/api/v1/auth/tokens/:id/", wrapper.DestroyToken)
	router.GET(baseURL+"/api/v1/auth/tokens/:id/", wrapper.RetrieveToken)
	router.PATCH(baseURL+"/api/v1/auth/tokens/:id/", wrapper.PartialUpdateToken)
	router.PUT(baseURL+"/api/v1/auth/tokens/:id/", wrapper.UpdateToken)
	router.POST(baseURL+"/api/v1/captcha/", wrapper.CreateCaptcha)
	router.GET(baseURL+"/api/v1/domains/", wrapper.ListDomains)
	router.POST(baseURL+"/api/v1/domains/", wrapper.CreateDomain)
	router.DELETE(baseURL+"/api/v1/domains/:name/", wrapper.DestroyDomain)
	router.GET(baseURL+"/api/v1/domains/:name/", wrapper.RetrieveDomain)
	router.GET(baseURL+"/api/v1/domains/:name/rrsets/", wrapper.ListRRsets)
	router.PATCH(baseURL+"/api/v1/domains/:name/rrsets/", wrapper.PartialUpdateRRset)
	router.POST(baseURL+"/api/v1/domains/:name/rrsets/", wrapper.CreateRRset)
	router.PUT(baseURL+"/api/v1/domains/:name/rrsets/", wrapper.UpdateRRset)
	router.DELETE(baseURL+"/api/v1/domains/:name/rrsets/@/:type/", wrapper.DestroyApexRRset)
	router.GET(baseURL+"/api/v1/domains/:name/rrsets/@/:type/", wrapper.RetrieveApexRRset)
	router.PATCH(baseURL+"/api/v1/domains/:name/rrsets/@/:type/", wrapper.PartialUpdateApexRRset)
	router.PUT(baseURL+"/api/v1/domains/:name/rrsets/@/:type/", wrapper.UpdateApexRRset)
	router.DELETE(baseURL+"/api/v1/domains/:name/rrsets/:subname/:type/", wrapper.DestroySubnameRRset)
	router.GET(baseURL+"/api/v1/domains/:name/rrsets/:subname/:type/", wrapper.RetrieveSubnameRRset)
	router.PATCH(baseURL+"/api/v1/domains/:name/rrsets/:subname/:type/", wrapper.PartialUpdateSubnameRRset)
	router.PUT(baseURL+"/api/v1/domains/:name/rrsets/:subname/:type/", wrapper.UpdateSubnameRRset)
	router.POST(baseURL+"/api/v1/donation/", wrapper.CreateDonation)
	router.GET(baseURL+"/api/v1/dyndns/update", wrapper.ListDyndnsRRsets)
	router.GET(baseURL+"/api/v1/serials/", wrapper.ListSerials)
	router.GET(baseURL+"/api/v1/v/activate-account/:code/", wrapper.RetrieveAuthenticatedActivateUserAction)
	router.POST(baseURL+"/api/v1/v/activate-account/:code/", wrapper.CreateAuthenticatedActivateUserAction)
	router.GET(baseURL+"/api/v1/v/change-email/:code/", wrapper.RetrieveAuthenticatedChangeEmailUserAction)
	router.POST(baseURL+"/api/v1/v/change-email/:code/", wrapper.CreateAuthenticatedChangeEmailUserAction)
	router.GET(baseURL+"/api/v1/v/delete-account/:code/", wrapper.RetrieveAuthenticatedDeleteUserAction)
	router.POST(baseURL+"/api/v1/v/delete-account/:code/", wrapper.CreateAuthenticatedDeleteUserAction)
	router.GET(baseURL+"/api/v1/v/renew-domain/:code/", wrapper.RetrieveAuthenticatedRenewDomainBasicUserAction)
	router.POST(baseURL+"/api/v1/v/renew-domain/:code/", wrapper.CreateAuthenticatedRenewDomainBasicUserAction)
	router.GET(baseURL+"/api/v1/v/reset-password/:code/", wrapper.RetrieveAuthenticatedResetPasswordUserAction)
	router.POST(baseURL+"/api/v1/v/reset-password/:code/", wrapper.CreateAuthenticatedResetPasswordUserAction)
}
