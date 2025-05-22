package api

import (
	"context"
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

	// Check if the email is already in use
	existingUser, err := s.engine.LookupUserByEmail(r.Context(), string(req.Email))
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if existingUser != nil {
		_ = render.Render(w, r, api_utils.ErrConflict)
		return
	}

	// Afterwards create the user
	user := &store.User{
		ID:           ID,
		Email:        string(req.Email),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: passwordHash,
		Status:       store.StatusPending,
		Role:         store.RoleUser,
	}
	err = s.engine.SetUser(r.Context(), user)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	// Send event to verify the email address
	err = s.sendVerifyAccountEvent(r.Context(), ID, string(req.Email), req.FirstName, req.LastName)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

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
	// Check if the user exists
	existingUser, err := s.engine.LookupUser(r.Context(), ID)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if existingUser == nil {
		_ = render.Render(w, r, api_utils.ErrNotFound)
		return
	}

	err = s.engine.DeleteUser(r.Context(), ID)
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
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	user, err := s.engine.LookupUser(r.Context(), userID)
	if err != nil || user == nil {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
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
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	user, err := s.engine.LookupUser(r.Context(), userID)
	if err != nil || user == nil {
		_ = render.Render(w, r, api_utils.ErrUnauthorized)
		return
	}

	// Afterwards update the user
	req := new(UserUpdateCurrent)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	// Check if the current password is correct
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

func (s *Server) sendVerifyAccountEvent(ctx context.Context, userID uuid.UUID, email, firstName, lastName string) error {
	token, _, err := s.jwsSigner.CreateVerifyAccountToken(userID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(transport.VerifyAccountEvent{
		Recipient: email,
		Channel:   "email",
		FirstName: firstName,
		LastName:  lastName,
		Token:     token,
	})
	if err != nil {
		return err
	}
	return s.producer.Produce(ctx, transport.VerifyAccountTopic, &transport.Message{
		ID:   uuid.New().String(),
		Data: data,
	})
}
