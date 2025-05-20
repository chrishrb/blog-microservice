package api

import (
	"errors"
	"net/http"

	"github.com/chrishrb/blog-microservice/internal/api_utils"
	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/user-service/service"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	req := new(UserCreate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	ID := uuid.New()
	passwordHash, err := service.HashPassword(req.Password)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(errors.New("failed to hash password")))
		return
	}

	user := &store.User{
		ID:           ID,
		Email:        string(req.Email),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: passwordHash,
		// TODO: after mail verification set status here to PENDING
		Status: store.StatusActive,
		Role:   store.RoleUser,
	}
	err = s.engine.SetUser(r.Context(), user)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	// TODO: Send mail to verify the email address

	render.Status(r, http.StatusCreated)
	render.Render(w, r, &User{
		Id:        ID,
		Email:     openapi_types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      UserRole(user.Role),
		Status:    UserStatus(user.Status),
	})
}

func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
	offset, limit := api_utils.GetPaginationWithDefaults(params.Offset, params.Limit)

	users, err := s.engine.ListUsers(r.Context(), offset, limit)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	res := make([]render.Renderer, len(users))
	for i, user := range users {
		res[i] = &User{
			Id:        user.ID,
			Email:     openapi_types.Email(user.Email),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      UserRole(user.Role),
			Status:    UserStatus(user.Status),
		}
	}

	render.RenderList(w, r, res)
}

func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request, ID openapi_types.UUID) {
	err := s.engine.DeleteUser(r.Context(), ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) LookupUser(w http.ResponseWriter, r *http.Request, ID openapi_types.UUID) {
	user, err := s.engine.LookupUser(r.Context(), ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if user == nil {
		_ = render.Render(w, r, api_utils.ErrNotFound)
		return
	}

	render.Render(w, r, &User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      UserRole(user.Role),
		Status:    UserStatus(user.Status),
	})
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request, ID openapi_types.UUID) {
	// Check if the user exists
	user, err := s.engine.LookupUser(r.Context(), ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if user == nil {
		_ = render.Render(w, r, api_utils.ErrNotFound)
		return
	}

	// Afterwards update the user
	req := new(UserUpdate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}
	// TODO: Send email verification if email is updated
	if req.Email != nil {
		user.Email = string(*req.Email)
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Role != nil {
		user.Role = string(*req.Role)
	}
	if req.Status != nil {
		user.Status = string(*req.Status)
	}

	err = s.engine.SetUser(r.Context(), user)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Render(w, r, &User{
		Id:        ID,
		Email:     openapi_types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      UserRole(user.Role),
		Status:    UserStatus(user.Status),
	})
}

func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	user, err := s.engine.LookupUser(r.Context(), userID)
	if err != nil || user == nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(errors.New("user lookup error")))
		return
	}

	render.Render(w, r, &User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      UserRole(user.Role),
		Status:    UserStatus(user.Status),
	})
}

func (s *Server) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	user, err := s.engine.LookupUser(r.Context(), userID)
	if err != nil || user == nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(errors.New("user lookup error")))
		return
	}

	render.Render(w, r, &User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      UserRole(user.Role),
		Status:    UserStatus(user.Status),
	})

	// Afterwards update the user
	req := new(UserUpdateCurrent)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	if ok := service.VerifyPassword(req.CurrentPassword, user.PasswordHash); !ok {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(errors.New("current password is incorrect")))
		return
	}

	// TODO: Send email verification if email is updated
	if req.Email != nil {
		user.Email = string(*req.Email)
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Password != nil {
		user.PasswordHash, err = service.HashPassword(*req.Password)
		if err != nil {
			_ = render.Render(w, r, api_utils.ErrInternalError(errors.New("failed to hash password")))
			return
		}
	}

	err = s.engine.SetUser(r.Context(), user)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Render(w, r, &User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      UserRole(user.Role),
		Status:    UserStatus(user.Status),
	})
}
