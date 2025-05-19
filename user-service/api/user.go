package api

import (
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *Server) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
	// TODO: Implement list users logic
}

func (h *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement user creation logic
}

func (h *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement current user retrieval logic
}

func (h *Server) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement current user update logic
}

func (h *Server) DeleteUser(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID) {
	// TODO: Implement user deletion logic
}

func (h *Server) LookupUserById(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID) {
	// TODO: Implement user lookup logic
}

func (h *Server) UpdateUser(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID) {
	// TODO: Implement user update logic
}
