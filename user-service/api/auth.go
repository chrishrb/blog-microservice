package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/chrishrb/blog-microservice/internal/api_utils"
	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/user-service/service"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

var InvalidTokenErr = errors.New("invalid or expired token")

func (s *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
	req := new(LoginRequest)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	user, err := s.engine.LookupUserByEmail(r.Context(), string(req.Email))
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	// User not found or not active
	if user == nil || user.Status != store.StatusActive {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	// Wrong password
	if !service.VerifyPassword(req.Password, user.PasswordHash) {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	claims := make([]string, 0)
	if user.Role == store.RoleAdmin {
		claims = append(claims, "all-users:read", "all-users:write")
	}

	accessToken, accessTokenExpiresIn, err := s.jwsSigner.CreateAccessToken(user.ID, claims)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	refreshToken, refreshTokenExpiresIn, err := s.jwsSigner.CreateRefreshToken(user.ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	err = s.engine.SetToken(r.Context(), &store.Token{
		UserID:  user.ID,
		Token:   refreshToken,
		TTL:     refreshTokenExpiresIn,
		Revoked: false,
	})
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Render(w, r, &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenExpiresIn.Seconds()),
	})
}

func (s *Server) LogoutUser(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	err = s.engine.SetTokenRevoked(r.Context(), userID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) RefreshToken(w http.ResponseWriter, r *http.Request) {
	req := new(RefreshTokenRequest)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	jwt, err := s.jwsVerifier.ValidateToken(req.RefreshToken)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	tokenRevoked, err := s.engine.IsTokenRevoked(r.Context(), req.RefreshToken)
	if err != nil || tokenRevoked {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	userID, err := auth.GetUserIDFromToken(jwt)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	user, err := s.engine.LookupUser(r.Context(), userID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	// User not found or not active
	if user == nil || user.Status != store.StatusActive {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	claims := make([]string, 0)
	if user.Role == store.RoleAdmin {
		claims = append(claims, "all-users:read", "all-users:write")
	}

	accessToken, accessTokenExpiresIn, err := s.jwsSigner.CreateAccessToken(user.ID, claims)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	refreshToken, refreshTokenExpiresIn, err := s.jwsSigner.CreateRefreshToken(user.ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	err = s.engine.SetToken(r.Context(), &store.Token{
		UserID:  user.ID,
		Token:   refreshToken,
		TTL:     refreshTokenExpiresIn,
		Revoked: false,
	})
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Render(w, r, &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenExpiresIn.Seconds()),
	})
}

func (s *Server) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	req := new(PasswordResetRequest)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	user, err := s.engine.LookupUserByEmail(r.Context(), string(req.Email))
	if err != nil || user == nil || user.Status != store.StatusActive {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(InvalidTokenErr))
		return
	}

	token, _, err := s.jwsSigner.CreatePasswordResetToken(user.ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	data, err := json.Marshal(transport.PasswordResetEvent{
		Recipient: string(req.Email),
		Channel:   "email",
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Token:     token,
	})
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	err = s.producer.Produce(r.Context(), transport.PasswordResetTopic, &transport.Message{
		ID:   uuid.New().String(),
		Data: data,
	})
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
}

func (s *Server) ResetPassword(w http.ResponseWriter, r *http.Request, token string) {
	// Validate request
	req := new(PasswordResetConfirmation)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(errors.New("passwords do not match")))
		return
	}

	// Validate the token and user
	tokenData, err := s.jwsVerifier.ValidateToken(token)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(InvalidTokenErr))
		return
	}
	userID, err := auth.GetUserIDFromToken(tokenData)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(InvalidTokenErr))
		return
	}
	user, err := s.engine.LookupUser(r.Context(), userID)
	if err != nil || user == nil || user.Status != store.StatusActive {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(InvalidTokenErr))
		return
	}

	// Update the user's password
	user.PasswordHash, err = service.HashPassword(req.NewPassword)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	err = s.engine.SetUser(r.Context(), user)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
}
