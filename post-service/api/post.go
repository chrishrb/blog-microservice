package api

import (
	"net/http"

	"github.com/chrishrb/blog-microservice/internal/api_utils"
	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	req := new(PostCreate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	ID := uuid.New()
	post := &store.Post{
		ID:       ID,
		AuthorID: userID,
		Title:    req.Title,
		Content:  req.Content,
	}
	if req.Tags != nil {
		post.Tags = *req.Tags
	}
	if req.Published != nil {
		post.Published = *req.Published
	} else {
		post.Published = false
	}
	err = s.engine.SetPost(r.Context(), post)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	_ = render.Render(w, r, &Post{
		Id:        ID,
		AuthorId:  post.AuthorID,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      &post.Tags,
		Published: post.Published,
	})
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	err := s.engine.DeletePost(r.Context(), id)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) LookupPost(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	post, err := s.engine.LookupPost(r.Context(), id)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if post == nil {
		_ = render.Render(w, r, api_utils.ErrNotFound)
		return
	}

	_ = render.Render(w, r, &Post{
		Id:        post.ID,
		AuthorId:  post.AuthorID,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      &post.Tags,
		Published: post.Published,
	})
}

func (s *Server) UpdatePost(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	// Check if the post exists
	post, err := s.engine.LookupPost(r.Context(), id)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}
	if post == nil {
		_ = render.Render(w, r, api_utils.ErrNotFound)
		return
	}

	// Afterwards update the post
	req := new(PostUpdate)
	if err := render.Bind(r, req); err != nil {
		_ = render.Render(w, r, api_utils.ErrInvalidRequest(err))
		return
	}

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.Tags != nil {
		post.Tags = *req.Tags
	}
	if req.Published != nil {
		post.Published = *req.Published
	}

	err = s.engine.SetPost(r.Context(), post)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	_ = render.Render(w, r, &Post{
		Id:        post.ID,
		AuthorId:  post.AuthorID,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      &post.Tags,
		Published: post.Published,
	})
}

func (s *Server) ListPosts(w http.ResponseWriter, r *http.Request, params ListPostsParams) {
	offset, limit := api_utils.GetPaginationWithDefaults(params.Offset, params.Limit)

	posts, err := s.engine.ListPosts(r.Context(), offset, limit)
	if err != nil {
		_ = render.Render(w, r, api_utils.ErrInternalError(err))
		return
	}

	res := make([]render.Renderer, len(posts))
	for i, p := range posts {
		res[i] = &Post{
			Id:        p.ID,
			AuthorId:  p.AuthorID,
			Title:     p.Title,
			Content:   p.Content,
			Published: p.Published,
		}
	}

	_ = render.RenderList(w, r, res)
}
