package auth

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/chrishrb/blog-microservice/internal/source"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/ecdsafile"
)

type JWSVerifier interface {
	ValidateToken(jws string) (jwt.Token, error)
	ValidatePasswordResetToken(jws string) (jwt.Token, error)
}

type LocalJWSVerifier struct {
	publicKey *ecdsa.PublicKey
	issuer    string
	audience  string
}

func NewLocalJWSVerifier(publicKeySource source.SourceProvider, issuer, audience string) (*LocalJWSVerifier, error) {
	publicKey, err := publicKeySource.GetData()
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	pubKey, err := ecdsafile.LoadEcdsaPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("loading PEM public key: %w", err)
	}

	return &LocalJWSVerifier{
		publicKey: pubKey,
		issuer:    issuer,
		audience:  audience,
	}, nil
}

// ValidateToken ensures that the critical JWT claims needed to ensure that we
// trust the JWT are present and with the correct values.
func (v *LocalJWSVerifier) ValidateToken(jwsString string) (jwt.Token, error) {
	return jwt.Parse(
		[]byte(jwsString),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithVerify(jwa.ES256, v.publicKey),
	)
}

// ValidatePasswordToken ensures that the critical JWT claims needed to ensure that we
// trust the JWT are present and with the correct values.
func (v *LocalJWSVerifier) ValidatePasswordResetToken(jwsString string) (jwt.Token, error) {
	return jwt.Parse(
		[]byte(jwsString),
		jwt.WithClaimValue(TypeClaim, TypePasswordReset),
		jwt.WithVerify(jwa.ES256, v.publicKey),
	)
}
