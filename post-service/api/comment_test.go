package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrishrb/blog-microservice/internal/testutil"
	"github.com/chrishrb/blog-microservice/post-service/api"
	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateComment(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	postID := uuid.New()

	// Create a post first
	post := &store.Post{
		ID:        postID,
		AuthorID:  userID,
		Title:     "Test Post",
		Content:   "Test Content",
		Tags:      []string{"test"},
		Published: true,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Create a comment
	d := api.CommentCreate{
		Content: "Test comment content",
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/posts/%s/comments", postID),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	var res api.Comment
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.NotEmpty(t, res.Id)
	assert.Equal(t, userID, res.AuthorId)
	assert.Equal(t, "Test comment content", res.Content)

	// Check the database
	dbComment, err := engine.LookupComment(req.Context(), postID, res.Id)
	require.NoError(t, err)
	assert.Equal(t, res.Id, dbComment.ID)
	assert.Equal(t, userID, dbComment.AuthorID)
	assert.Equal(t, postID, dbComment.PostID)
	assert.Equal(t, "Test comment content", dbComment.Content)
}

func TestDeleteComment(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	postID := uuid.New()
	commentID := uuid.New()

	// Create a post first
	post := &store.Post{
		ID:        postID,
		AuthorID:  userID,
		Title:     "Test Post",
		Content:   "Test Content",
		Tags:      []string{"test"},
		Published: true,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Create a comment
	comment := &store.Comment{
		ID:        commentID,
		AuthorID:  userID,
		PostID:    postID,
		Content:   "Test comment content",
	}
	err = engine.SetComment(t.Context(), comment)
	require.NoError(t, err)

	// Delete the comment
	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/posts/%s/comments/%s", postID, commentID),
		nil,
	)
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	// Verify comment no longer exists
	dbComment, err := engine.LookupComment(req.Context(), postID, commentID)
	require.NoError(t, err)
	assert.Nil(t, dbComment)
}

func TestLookupComment(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	postID := uuid.New()
	commentID := uuid.New()

	// Create a post first
	post := &store.Post{
		ID:        postID,
		AuthorID:  userID,
		Title:     "Test Post",
		Content:   "Test Content",
		Tags:      []string{"test"},
		Published: true,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Create a comment
	comment := &store.Comment{
		ID:        commentID,
		AuthorID:  userID,
		PostID:    postID,
		Content:   "Test comment content",
	}
	err = engine.SetComment(t.Context(), comment)
	require.NoError(t, err)

	// Lookup the comment
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/posts/%s/comments/%s", postID, commentID),
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.Comment
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, commentID, res.Id)
	assert.Equal(t, userID, res.AuthorId)
	assert.Equal(t, "Test comment content", res.Content)
}

func TestLookupComment_NotFound(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	postID := uuid.New()
	commentID := uuid.New()

	// Lookup a non-existent comment
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/posts/%s/comments/%s", postID, commentID),
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestUpdateComment(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	postID := uuid.New()
	commentID := uuid.New()

	// Create a post first
	post := &store.Post{
		ID:        postID,
		AuthorID:  userID,
		Title:     "Test Post",
		Content:   "Test Content",
		Tags:      []string{"test"},
		Published: true,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Create a comment
	comment := &store.Comment{
		ID:        commentID,
		AuthorID:  userID,
		PostID:    postID,
		Content:   "Original comment content",
	}
	err = engine.SetComment(t.Context(), comment)
	require.NoError(t, err)

	// Update the comment
	d := api.CommentUpdate{
		Content: testutil.Ptr("Updated comment content"),
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/posts/%s/comments/%s", postID, commentID),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.Comment
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, commentID, res.Id)
	assert.Equal(t, userID, res.AuthorId)
	assert.Equal(t, "Updated comment content", res.Content)

	// Check the database
	dbComment, err := engine.LookupComment(req.Context(), postID, commentID)
	require.NoError(t, err)
	assert.Equal(t, res.Id, dbComment.ID)
	assert.Equal(t, userID, dbComment.AuthorID)
	assert.Equal(t, postID, dbComment.PostID)
	assert.Equal(t, "Updated comment content", dbComment.Content)
}

func TestUpdateComment_NotFound(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	postID := uuid.New()
	commentID := uuid.New()

	d := api.CommentUpdate{
		Content: testutil.Ptr("Updated comment content"),
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/posts/%s/comments/%s", postID, commentID),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestListComments(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	postID := uuid.New()

	// Create a post first
	post := &store.Post{
		ID:        postID,
		AuthorID:  userID,
		Title:     "Test Post",
		Content:   "Test Content",
		Tags:      []string{"test"},
		Published: true,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Create multiple comments
	commentID1 := uuid.New()
	comment1 := &store.Comment{
		ID:        commentID1,
		AuthorID:  userID,
		PostID:    postID,
		Content:   "First comment",
	}
	err = engine.SetComment(t.Context(), comment1)
	require.NoError(t, err)

	commentID2 := uuid.New()
	comment2 := &store.Comment{
		ID:        commentID2,
		AuthorID:  userID,
		PostID:    postID,
		Content:   "Second comment",
	}
	err = engine.SetComment(t.Context(), comment2)
	require.NoError(t, err)

	// List comments
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/posts/%s/comments", postID),
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var resList []api.Comment
	err = json.NewDecoder(rr.Body).Decode(&resList)
	require.NoError(t, err)
	require.Equal(t, 2, len(resList))

	// Check that our created comments exist in the response
	var ids []uuid.UUID
	for _, comment := range resList {
		ids = append(ids, comment.Id)
	}
	assert.Contains(t, ids, commentID1)
	assert.Contains(t, ids, commentID2)
}

