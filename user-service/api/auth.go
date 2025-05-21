package api

import (
	"net/http"

	"github.com/chrishrb/blog-microservice/internal/api_utils"
	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/user-service/service"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/go-chi/render"
)

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

	accessToken, err := s.JWSSigner.CreateJWS(user.ID, claims)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	refreshToken, err := s.JWSSigner.CreateRefreshJWS(user.ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Render(w, r, &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.JWSSigner.GetAccessTokenExpiresIn().Seconds()),
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

	jwt, err := s.JWSVerifier.ValidateJWS(req.RefreshToken)
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

	accessToken, err := s.JWSSigner.CreateJWS(user.ID, claims)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	refreshToken, err := s.JWSSigner.CreateRefreshJWS(user.ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	s.engine.SetToken(r.Context(), &store.Token{
		UserID:  user.ID,
		Token:   refreshToken,
		TTL:     s.JWSSigner.GetRefreshTokenExpiresIn(),
		Revoked: false,
	})

	render.Render(w, r, &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.JWSSigner.GetAccessTokenExpiresIn().Seconds()),
	})
}

func (s *Server) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement password reset request logic
}

func (s *Server) ResetPassword(w http.ResponseWriter, r *http.Request, token string) {
	// TODO: Implement password reset logic
}
