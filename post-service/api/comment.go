package api

import (
	"net/http"

	"github.com/chrishrb/blog-microservice/internal/api_utils"
	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (s *Server) CreateComment(w http.ResponseWriter, r *http.Request, postId uuid.UUID) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	req := new(CommentCreate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	ID := uuid.New()
	comment := &store.Comment{
		ID:       ID,
		AuthorID: userID,
		PostID:   postId,
		Content:  req.Content,
	}
	err = s.engine.SetComment(r.Context(), comment)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, &Comment{
		Id:       ID,
		AuthorId: comment.AuthorID,
		Content:  comment.Content,
	})
}

func (s *Server) ListComments(w http.ResponseWriter, r *http.Request, postId uuid.UUID, params ListCommentsParams) {
	offset, limit := api_utils.GetPaginationWithDefaults(params.Offset, params.Limit)

	comment, err := s.engine.ListCommentsByPostID(r.Context(), postId, offset, limit)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	comments := make([]render.Renderer, len(comment))
	for i, c := range comment {
		comments[i] = &Comment{
			Id:       c.ID,
			AuthorId: c.AuthorID,
			Content:  c.Content,
		}
	}

	render.RenderList(w, r, comments)
}

func (s *Server) LookupComment(w http.ResponseWriter, r *http.Request, postId, id uuid.UUID) {
	comment, err := s.engine.LookupComment(r.Context(), postId, id)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if comment == nil {
		_ = render.Render(w, r, api_utils.ErrNotFound)
		return
	}

	render.Render(w, r, &Comment{
		Id:       comment.ID,
		AuthorId: comment.AuthorID,
		Content:  comment.Content,
	})
}

func (s *Server) UpdateComment(w http.ResponseWriter, r *http.Request, postId, id uuid.UUID) {
	// Check if the comment exists
	comment, err := s.engine.LookupComment(r.Context(), postId, id)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if comment == nil {
		_ = render.Render(w, r, api_utils.ErrNotFound)
		return
	}

	// Afterwards update the comment
	req := new(CommentUpdate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	if req.Content != nil {
		comment.Content = *req.Content
	}

	err = s.engine.SetComment(r.Context(), comment)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Render(w, r, &Comment{
		Id:       comment.ID,
		AuthorId: comment.AuthorID,
		Content:  comment.Content,
	})
}

func (s *Server) DeleteComment(w http.ResponseWriter, r *http.Request, postId, id uuid.UUID) {
	err := s.engine.DeleteComment(r.Context(), postId, id)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
