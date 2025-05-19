package api

import (
	"net/http"

	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (s *Server) CreateComment(w http.ResponseWriter, r *http.Request, postId string) {
	req := new(CommentCreate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	ID := uuid.New().String()
	comment := &store.Comment{
		ID:       ID,
		AuthorID: req.AuthorId,
		PostID:   postId,
		Content:  req.Content,
	}
	err := s.engine.SetComment(r.Context(), comment)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, &Comment{
		Id:       ID,
		AuthorId: comment.AuthorID,
		Content:  comment.Content,
	})
}

func (s *Server) ListCommentsByPostId(w http.ResponseWriter, r *http.Request, postId string) {
	comment, err := s.engine.ListCommentsByPostID(r.Context(), postId)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
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

func (s *Server) LookupComment(w http.ResponseWriter, r *http.Request, postId, id string) {
	comment, err := s.engine.LookupComment(r.Context(), postId, id)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
	if comment == nil {
		_ = render.Render(w, r, ErrNotFound)
		return
	}

	render.Render(w, r, &Comment{
		Id:       comment.ID,
		AuthorId: comment.AuthorID,
		Content:  comment.Content,
	})
}

func (s *Server) UpdateComment(w http.ResponseWriter, r *http.Request, postId, id string) {
	// Check if the comment exists
	comment, err := s.engine.LookupComment(r.Context(), postId, id)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
	if comment == nil {
		_ = render.Render(w, r, ErrNotFound)
		return
	}

	// Afterwards update the comment
	req := new(CommentUpdate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if req.AuthorId != nil {
		comment.AuthorID = *req.AuthorId
	}
	if req.Content != nil {
		comment.Content = *req.Content
	}

	err = s.engine.SetComment(r.Context(), comment)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}

	render.Render(w, r, &Comment{
		Id:       comment.ID,
		AuthorId: comment.AuthorID,
		Content:  comment.Content,
	})
}

func (s *Server) DeleteComment(w http.ResponseWriter, r *http.Request, postId, id string) {
	err := s.engine.DeleteComment(r.Context(), postId, id)
	if err != nil {
		_ = render.Render(w, r, ErrInternalError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
