package main

import (
	"errors"
)

var (
	ErrUsedSystemInfo                   = errors.New("used systeminfo")
	ErrUsedServiceInfo                  = errors.New("used serviceinfo")
	ErrUsedModuleInfo                   = errors.New("used moduleinfo")
	ErrSubmitSystemInfo                 = errors.New("submit systeminfo")
	ErrSubmitServiceInfo                = errors.New("submit serviceinfo")
	ErrSubmitModuleInfo                 = errors.New("submit moduleinfo")
	ErrNotForwardVersion                = errors.New("not forward version")
	ErrNotSystemInfoOwner               = errors.New("not systeminfo owner")
	ErrNotServiceInfoOwner              = errors.New("not serviceinfo owner")
	ErrNotExistServiceClaim             = errors.New("not exist service claim")
	ErrNotAuthorized                    = errors.New("not authorized")
	ErrExistMember                      = errors.New("exist member")
	ErrExistUser                        = errors.New("exist user")
	ErrExistRepoName                    = errors.New("exist repo name")
	ErrNotExistRepo                     = errors.New("not exist repo")
	ErrUnknownSamlResponse              = errors.New("unknown saml response")
	ErrNotExistResetToken               = errors.New("not exist reset token")
	ErrInvalidResetToken                = errors.New("invalid token")
	ErrInvalidGitAuth                   = errors.New("invalid git auth")
	ErrNotExistAPIDocRouteIndex         = errors.New("not exist apidoc route index")
	ErrNotExistAPIDocRouteParamterIndex = errors.New("not exist apidoc route parameter index")
)
