package shared

import (
	"context"
	"net/http"

	"github.com/victorvbello/gomcp/mcp/types"
)

type AUTH_INFO_REQUEST_KEY_TYPE string

const AUTH_INFO_REQUEST_KEY_NAME = AUTH_INFO_REQUEST_KEY_TYPE("authInfo")

func MakeAuthInfoRequest(req *http.Request, authInfo types.AuthInfo) context.Context {
	ctx := context.WithValue(req.Context(), AUTH_INFO_REQUEST_KEY_NAME, authInfo)
	return ctx
}

func GetAuthInfoRequest(r *http.Request) types.AuthInfo {
	value := r.Context().Value(AUTH_INFO_REQUEST_KEY_NAME)
	return value.(types.AuthInfo)
}
