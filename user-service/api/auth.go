package api

import (
	"net/http"

	"github.com/chrishrb/blog-microservice/internal/api_utils"
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

	var claims []string
	if user.Role == store.RoleAdmin {
		claims = append(claims, "all-users:read", "all-users:write")
	}

	token, expiresIn, err := s.authService.CreateJWSWithClaims(claims)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Render(w, r, &AuthResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
	})
}

func (s *Server) LogoutUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logout logic
}

func (s *Server) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement password reset request logic
}

func (s *Server) ResetPassword(w http.ResponseWriter, r *http.Request, token string) {
	// TODO: Implement password reset logic
}
