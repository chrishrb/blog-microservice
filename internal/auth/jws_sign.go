package auth

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/chrishrb/blog-microservice/internal/source"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/ecdsafile"
)

const PermissionsClaim = "permissions"
const UserIDClaim = "user_id"

type JWSSigner interface {
	CreateJWS(userID uuid.UUID, claims []string) (string, error)
	GetExpiresIn() time.Duration
}

type LocalJWSSigner struct {
	// PrivateKey is an ECDSA private key which was generated with the following
	// command:
	//
	//	openssl ecparam -name prime256v1 -genkey -noout -out ecprivatekey.pem
	privateKey *ecdsa.PrivateKey
	issuer     string
	audience   string
	expiresIn  time.Duration
}

func NewLocalJWSSigner(privateKeySource source.SourceProvider, issuer, audience string, expiresIn time.Duration) (*LocalJWSSigner, error) {
	privateKey, err := privateKeySource.GetData()
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	p, err := ecdsafile.LoadEcdsaPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("loading PEM private key: %w", err)
	}

	return &LocalJWSSigner{
		privateKey: p,
		issuer:     issuer,
		audience:   audience,
		expiresIn:  expiresIn,
	}, nil
}

// CreateJWS creates a JWS with the given user ID and claims. The claims are
// added to the "permissions" claim in the JWS.
func (s *LocalJWSSigner) CreateJWS(userID uuid.UUID, claims []string) (string, error) {
	t := jwt.New()
	err := t.Set(jwt.IssuerKey, s.issuer)
	if err != nil {
		return "", fmt.Errorf("setting issuer: %w", err)
	}
	err = t.Set(jwt.AudienceKey, s.audience)
	if err != nil {
		return "", fmt.Errorf("setting audience: %w", err)
	}
	err = t.Set(UserIDClaim, userID.String())
	if err != nil {
		return "", fmt.Errorf("setting userID: %w", err)
	}
	err = t.Set(PermissionsClaim, claims)
	if err != nil {
		return "", fmt.Errorf("setting permissions: %w", err)
	}
	err = t.Set(jwt.ExpirationKey, time.Now().Add(s.expiresIn).Unix())
	if err != nil {
		return "", fmt.Errorf("setting expiration: %w", err)
	}
	token, err := s.signToken(t)
	if err != nil {
		return "", err
	}
	return string(token),  nil
}

func (s *LocalJWSSigner) GetExpiresIn() time.Duration {
	return s.expiresIn
}

// SignToken takes a JWT and signs it with our private key, returning a JWS.
func (s *LocalJWSSigner) signToken(t jwt.Token) ([]byte, error) {
	hdr := jws.NewHeaders()
	if err := hdr.Set(jws.AlgorithmKey, jwa.ES256); err != nil {
		return nil, fmt.Errorf("setting algorithm: %w", err)
	}
	if err := hdr.Set(jws.TypeKey, "JWT"); err != nil {
		return nil, fmt.Errorf("setting type: %w", err)
	}
	return jwt.Sign(t, jwa.ES256, s.privateKey, jwt.WithHeaders(hdr))
}
