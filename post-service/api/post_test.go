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

func TestCreatePost(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	d := api.PostCreate{
		Title:     "someTitle",
		Content:   "someContent",
		Published: testutil.Ptr(true),
		Tags:      &[]string{"tag1", "tag2"},
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/posts",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	var res api.Post
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.NotEmpty(t, res.Id)
	assert.Equal(t, userID, res.AuthorId)
	assert.Equal(t, "someTitle", res.Title)
	assert.Equal(t, "someContent", res.Content)
	assert.Equal(t, testutil.Ptr([]string{"tag1", "tag2"}), res.Tags)
	assert.True(t, res.Published)

	// Check the database
	dbPost, err := engine.LookupPost(req.Context(), res.Id)
	require.NoError(t, err)
	assert.Equal(t, res.Id, dbPost.ID)
	assert.Equal(t, userID, dbPost.AuthorID)
	assert.Equal(t, "someTitle", dbPost.Title)
	assert.Equal(t, "someContent", dbPost.Content)
	assert.Equal(t, []string{"tag1", "tag2"}, dbPost.Tags)
	assert.True(t, dbPost.Published)
}

func TestDeletePost(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create a post first
	ID := uuid.New()
	post := &store.Post{
		ID:        ID,
		AuthorID:  userID,
		Title:     "someTitle",
		Content:   "someContent",
		Tags:      []string{"tag1", "tag2"},
		Published: false,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Delete the post
	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/posts/%s", post.ID),
		nil,
	)
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	// Verify post no longer exists
	dbPost, err := engine.LookupPost(req.Context(), post.ID)
	require.NoError(t, err)
	assert.Nil(t, dbPost)
}

func TestLookupPost(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	ID := uuid.New()
	userID := uuid.New()

	// Create a post first
	post := &store.Post{
		ID:        ID,
		AuthorID:  userID,
		Title:     "someTitle",
		Content:   "someContent",
		Tags:      []string{"tag1", "tag2"},
		Published: false,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Lookup the post
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/posts/%s", post.ID),
		nil,
	)
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.Post
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, post.ID, res.Id)
	assert.Equal(t, userID, res.AuthorId)
	assert.Equal(t, "someTitle", res.Title)
	assert.Equal(t, "someContent", res.Content)
	assert.Equal(t, testutil.Ptr([]string{"tag1", "tag2"}), res.Tags)
	assert.False(t, res.Published)
}

func TestLookupPost_NotFound(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	// Lookup a non-existent post
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/posts/%s", uuid.New()),
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestUpdatePost(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	ID := uuid.New()
	userID := uuid.New()

	// Create a post first
	post := &store.Post{
		ID:        ID,
		AuthorID:  userID,
		Title:     "someTitle",
		Content:   "someContent",
		Tags:      []string{"tag1", "tag2"},
		Published: false,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	// Update the post
	d := api.PostUpdate{
		Title:     testutil.Ptr("Updated Title"),
		Content:   testutil.Ptr("Updated Content"),
		Published: testutil.Ptr(true),
		Tags:      testutil.Ptr([]string{"updated", "tags"}),
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/posts/%s", post.ID),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.Post
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, post.ID, res.Id)
	assert.Equal(t, "Updated Title", res.Title)
	assert.Equal(t, "Updated Content", res.Content)
	assert.Equal(t, testutil.Ptr([]string{"updated", "tags"}), res.Tags)
	assert.True(t, res.Published)

	// Check the database
	dbPost, err := engine.LookupPost(req.Context(), res.Id)
	require.NoError(t, err)
	assert.Equal(t, res.Id, dbPost.ID)
	assert.Equal(t, "Updated Title", dbPost.Title)
	assert.Equal(t, "Updated Content", dbPost.Content)
	assert.Equal(t, []string{"updated", "tags"}, dbPost.Tags)
	assert.True(t, dbPost.Published)
}

func TestUpdatePost_NotFound(t *testing.T) {
	server, r, _, _ := setupServer(t)
	defer server.Close()

	d := api.PostUpdate{
		Title: testutil.Ptr("Updated Title"),
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/posts/%s", uuid.New()),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestListPosts(t *testing.T) {
	server, r, engine, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create multiple posts
	ID1 := uuid.New()
	post1 := &store.Post{
		ID:        ID1,
		AuthorID:  userID,
		Title:     "someTitle",
		Content:   "someContent",
		Tags:      []string{"tag1", "tag2"},
		Published: false,
	}
	err := engine.SetPost(t.Context(), post1)
	require.NoError(t, err)

	ID2 := uuid.New()
	post2 := &store.Post{
		ID:        ID2,
		AuthorID:  userID,
		Title:     "anotherTitle",
		Content:   "anotherContent",
		Tags:      []string{"tag2", "tag3"},
		Published: false,
	}
	err = engine.SetPost(t.Context(), post2)
	require.NoError(t, err)

	// List posts
	req := httptest.NewRequest(
		http.MethodGet,
		"/posts",
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var resList []api.Post
	err = json.NewDecoder(rr.Body).Decode(&resList)

	require.NoError(t, err)
	require.Equal(t, len(resList), 2)

	// Check that our created posts exist in the response
	var ids []uuid.UUID
	for _, post := range resList {
		ids = append(ids, post.Id)
	}
	assert.Contains(t, ids, post1.ID)
	assert.Contains(t, ids, post2.ID)
}
