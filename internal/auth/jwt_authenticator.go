package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/chrishrb/blog-microservice/internal/api_utils"
	"github.com/chrishrb/blog-microservice/internal/writeablecontext"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"
)

const UserIDContextKey = "userID"

var (
	ErrNoAuthHeader      = errors.New("authorization header is missing")
	ErrInvalidAuthHeader = errors.New("authorization header is malformed")
	ErrClaimsInvalid     = errors.New("provided claims do not match expected scopes")
)

func NewAuthenticator(v JWSVerifier) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return Authenticate(v, ctx, input)
	}
}

// Authenticate uses the specified validator to ensure a JWT is valid, then makes
// sure that the claims provided by the JWT match the scopes as required in the API.
func Authenticate(v JWSVerifier, ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// Our security scheme is named BearerAuth, ensure this is the case
	if input.SecuritySchemeName != "BearerAuth" {
		return fmt.Errorf("security scheme %s != 'BearerAuth'", input.SecuritySchemeName)
	}

	// Now, we need to get the JWS from the request, to match the request expectations
	// against request contents.
	jws, err := getJWSFromRequest(input.RequestValidationInput.Request)
	if err != nil {
		return fmt.Errorf("getting jws: %w", err)
	}

	// if the JWS is valid, we have a JWT, which will contain a bunch of claims.
	token, err := v.ValidateToken(jws)
	if err != nil {
		return fmt.Errorf("validating JWS: %w", err)
	}

	// We've got a valid token now, and we can look into its claims to see whether
	// they match. Every single scope must be present in the claims.
	err = checkTokenClaims(input.Scopes, token)

	if err != nil {
		return fmt.Errorf("token claims don't match: %w", err)
	}

	// Set the property on the context so the handler is able to
	// access the claims data we generate in here.
	userID, err := GetUserIDFromToken(token)
	if err != nil {
		return fmt.Errorf("userID not in token: %w", err)
	}
	reqstore := writeablecontext.FromContext(input.RequestValidationInput.Request.Context())
	reqstore.Set(UserIDContextKey, userID.String())

	return nil
}

// GetUserIDFromContext retrieves the user ID from the context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDAny, isValid := writeablecontext.FromContext(ctx).Get(UserIDContextKey)
	if !isValid {
		return uuid.Nil, fmt.Errorf("userID not found in context")
	}

	userIDStr, isValid := userIDAny.(string)
	if !isValid {
		return uuid.Nil, fmt.Errorf("userID is not a string")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("userID is not a valid UUID: %w", err)
	}

	return userID, nil
}

// Get userID from token
func GetUserIDFromToken(t jwt.Token) (uuid.UUID, error) {
	userIDAny, found := t.Get(jwt.SubjectKey)
	if !found {
		return uuid.Nil, fmt.Errorf("user ID claim not found")
	}
	userIDStr, ok := userIDAny.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID claim is not a string")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("userID is not a valid UUID: %w", err)
	}
	return userID, nil
}

// GetAuthMiddleware returns a middleware for validating requests against the
// OpenAPI spec.
func GetAuthMiddleware(swagger *openapi3.T, v JWSVerifier) func(next http.Handler) http.Handler {
	return oapimiddleware.OapiRequestValidatorWithOptions(swagger, &oapimiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: NewAuthenticator(v),
		},
		ErrorHandlerWithOpts: func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request, opts oapimiddleware.ErrorHandlerOpts) {
			_ = render.Render(w, r, &api_utils.ErrResponse{
				Err:            err,
				HTTPStatusCode: opts.StatusCode,
				StatusText:     http.StatusText(opts.StatusCode),
				ErrorText:      err.Error(),
			})
		},
	})
}

// getJWSFromRequest extracts a JWS string from an Authorization: Bearer <jws> header
func getJWSFromRequest(req *http.Request) (string, error) {
	authHdr := req.Header.Get("Authorization")
	// Check for the Authorization header.
	if authHdr == "" {
		return "", ErrNoAuthHeader
	}
	// We expect a header value of the form "Bearer <token>", with 1 space after
	// Bearer, per spec.
	prefix := "Bearer "
	if !strings.HasPrefix(authHdr, prefix) {
		return "", ErrInvalidAuthHeader
	}
	return strings.TrimPrefix(authHdr, prefix), nil
}

// getClaimsFromToken returns a list of claims from the token. We store these
// as a list under the "perms" claim, short for permissions, to keep the token
// shorter.
func getClaimsFromToken(t jwt.Token) ([]string, error) {
	rawPerms, found := t.Get(PermissionsClaim)
	if !found {
		// If the perms aren't found, it means that the token has none, but it has
		// passed signature validation by now, so it's a valid token, so we return
		// the empty list.
		return make([]string, 0), nil
	}

	// rawPerms will be an untyped JSON list, so we need to convert it to
	// a string list.
	rawList, ok := rawPerms.([]any)
	if !ok {
		return nil, fmt.Errorf("'%s' claim is unexpected type'", PermissionsClaim)
	}

	claims := make([]string, len(rawList))

	for i, rawClaim := range rawList {
		var ok bool
		claims[i], ok = rawClaim.(string)
		if !ok {
			return nil, fmt.Errorf("%s[%d] is not a string", PermissionsClaim, i)
		}
	}
	return claims, nil
}

func checkTokenClaims(expectedClaims []string, t jwt.Token) error {
	claims, err := getClaimsFromToken(t)
	if err != nil {
		return fmt.Errorf("getting claims from token: %w", err)
	}
	// Put the claims into a map, for quick access.
	claimsMap := make(map[string]bool, len(claims))
	for _, c := range claims {
		claimsMap[c] = true
	}

	for _, e := range expectedClaims {
		if !claimsMap[e] {
			return ErrClaimsInvalid
		}
	}
	return nil
}
