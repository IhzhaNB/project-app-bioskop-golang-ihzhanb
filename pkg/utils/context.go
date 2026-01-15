package utils

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
	TokenKey  contextKey = "token"
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userIDVal := ctx.Value(UserIDKey)
	if userIDVal == nil {
		return uuid.Nil, false
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}

func GetRoleFromContext(ctx context.Context) (string, bool) {
	roleVal := ctx.Value(RoleKey)
	if roleVal == nil {
		return "", false
	}

	role, ok := roleVal.(string)
	return role, ok
}

func SetUserContext(ctx context.Context, userID uuid.UUID, role string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID.String())
	ctx = context.WithValue(ctx, RoleKey, role)
	return ctx
}

// GetTokenFromContext mendapatkan token dari context
func GetTokenFromContext(ctx context.Context) (string, bool) {
	tokenVal := ctx.Value(TokenKey)
	if tokenVal == nil {
		return "", false
	}

	token, ok := tokenVal.(string)
	return token, ok
}

// SetTokenContext menambahkan token ke context
func SetTokenContext(ctx context.Context, token string) context.Context {
	ctx = context.WithValue(ctx, TokenKey, token)
	return ctx
}
