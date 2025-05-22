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
const TypeClaim = "type"
const TypePasswordReset = "password_reset"

type JWSSigner interface {
	CreateAccessToken(userID uuid.UUID, claims []string) (string, time.Duration, error)
	CreateRefreshToken(userID uuid.UUID) (string, time.Duration, error)
	CreatePasswordResetToken(userID uuid.UUID) (string, time.Duration, error)
}

type LocalJWSSigner struct {
	// PrivateKey is an ECDSA private key which was generated with the following
	// command:
	//
	//	openssl ecparam -name prime256v1 -genkey -noout -out ecprivatekey.pem
	privateKey            *ecdsa.PrivateKey
	issuer                string
	audience              string
	accessTokenExpiresIn  time.Duration
	refreshTokenExpiresIn time.Duration
}

func NewLocalJWSSigner(
	privateKeySource source.SourceProvider,
	issuer,
	audience string,
	accessTokenExpiresIn time.Duration,
	refreshTokenExpiresIn time.Duration,
) (*LocalJWSSigner, error) {
	privateKey, err := privateKeySource.GetData()
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	p, err := ecdsafile.LoadEcdsaPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("loading PEM private key: %w", err)
	}

	return &LocalJWSSigner{
		privateKey:            p,
		issuer:                issuer,
		audience:              audience,
		accessTokenExpiresIn:  accessTokenExpiresIn,
		refreshTokenExpiresIn: refreshTokenExpiresIn,
	}, nil
}

// CreateAccessToken creates a JWS with the given user ID and claims. The claims are
// added to the "permissions" claim in the JWS.
func (s *LocalJWSSigner) CreateAccessToken(userID uuid.UUID, claims []string) (string, time.Duration, error) {
	t := jwt.New()
	err := t.Set(jwt.IssuerKey, s.issuer)
	if err != nil {
		return "", 0, fmt.Errorf("setting issuer: %w", err)
	}
	err = t.Set(jwt.AudienceKey, s.audience)
	if err != nil {
		return "", 0, fmt.Errorf("setting audience: %w", err)
	}
	err = t.Set(jwt.SubjectKey, userID.String())
	if err != nil {
		return "", 0, fmt.Errorf("setting subject: %w", err)
	}
	err = t.Set(PermissionsClaim, claims)
	if err != nil {
		return "", 0, fmt.Errorf("setting permissions: %w", err)
	}
	err = t.Set(jwt.ExpirationKey, time.Now().Add(s.accessTokenExpiresIn).Unix())
	if err != nil {
		return "", 0, fmt.Errorf("setting expiration: %w", err)
	}
	token, err := s.signToken(t)
	if err != nil {
		return "", 0, err
	}
	return string(token), s.accessTokenExpiresIn, nil
}

// CreateRefreshToken creates a refresh JWS
func (s *LocalJWSSigner) CreateRefreshToken(userID uuid.UUID) (string, time.Duration, error) {
	t := jwt.New()
	err := t.Set(jwt.IssuerKey, s.issuer)
	if err != nil {
		return "", 0, fmt.Errorf("setting issuer: %w", err)
	}
	err = t.Set(jwt.AudienceKey, s.audience)
	if err != nil {
		return "", 0, fmt.Errorf("setting audience: %w", err)
	}
	err = t.Set(jwt.SubjectKey, userID.String())
	if err != nil {
		return "", 0, fmt.Errorf("setting subject: %w", err)
	}
	err = t.Set(jwt.ExpirationKey, time.Now().Add(s.refreshTokenExpiresIn).Unix())
	if err != nil {
		return "", 0, fmt.Errorf("setting expiration: %w", err)
	}
	token, err := s.signToken(t)
	if err != nil {
		return "", 0, err
	}
	return string(token), s.refreshTokenExpiresIn, nil
}

// CreateRefreshToken creates a refresh JWS
func (s *LocalJWSSigner) CreatePasswordResetToken(userID uuid.UUID) (string, time.Duration, error) {
	t := jwt.New()
	err := t.Set(jwt.SubjectKey, userID.String())
	expiresIn := time.Duration(15 * time.Minute)
	if err != nil {
		return "", 0, fmt.Errorf("setting subject: %w", err)
	}
	err = t.Set(TypeClaim, TypePasswordReset)
	if err != nil {
		return "", 0, fmt.Errorf("setting subject: %w", err)
	}
	err = t.Set(jwt.ExpirationKey, time.Now().Add(expiresIn).Unix())
	if err != nil {
		return "", 0, fmt.Errorf("setting expiration: %w", err)
	}
	token, err := s.signToken(t)
	if err != nil {
		return "", 0, err
	}
	return string(token), expiresIn, nil
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
