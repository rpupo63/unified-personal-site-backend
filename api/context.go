package api

import (
	"context"
)

type keyType string

const (
	userIDKey         keyType = "userID"
	organizationIDKey keyType = "organizationID"
	userKey           keyType = "user"
)

// ctxWithUserID adds a user ID to the context
func ctxWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

/*
// ctxWithOrganizationID adds an organization ID to the context
func ctxWithOrganizationID(ctx context.Context, organizationID string) context.Context {
	return context.WithValue(ctx, organizationIDKey, organizationID)
}

// ctxGetUserID retrieves a user ID from the context
func ctxGetUserID(ctx context.Context) (string, error) {
	return ctxGetStringValue(ctx, userIDKey)
}


// ctxGetOrganizationID retrieves an organization ID from the context
func ctxGetOrganizationID(ctx context.Context) (string, error) {
	return ctxGetStringValue(ctx, organizationIDKey)
}

// ctxGetStringValue is a helper function to retrieve string values from the context by key
func ctxGetStringValue(ctx context.Context, key keyType) (string, error) {
	if ctxValue := ctx.Value(key); ctxValue == nil {
		return "", errors.New("key not found in context")
	} else if valueAsString, ok := ctxValue.(string); !ok {
		return "", errors.New("value is not of type `string`")
	} else {
		return valueAsString, nil
	}
}*/
