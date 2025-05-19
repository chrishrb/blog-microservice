package inmemory_test

import (
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/chrishrb/blog-microservice/post-service/store/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clock_testing "k8s.io/utils/clock/testing"
)

func TestSetComment(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	err := engine.SetPost(t.Context(), &store.Post{
		ID:        "1",
		AuthorID:  "someAuthorID",
		Title:     "Some Title",
		Content:   "Some Content",
		Tags:      []string{"tag1", "tag2"},
		Published: true,
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       "1",
		AuthorID: "someAuthorID",
		PostID:   "1",
		Content:  "Some Comment",
	})
	assert.NoError(t, err)

	post, err := engine.LookupPost(t.Context(), "1")
	assert.NoError(t, err)
	assert.Equal(t, "1", post.ID)
	assert.Equal(t, "someAuthorID", post.AuthorID)
	assert.Equal(t, "Some Title", post.Title)
	assert.Equal(t, "Some Content", post.Content)
	assert.Equal(t, []string{"tag1", "tag2"}, post.Tags)
	assert.True(t, post.Published)
	assert.Equal(t, fakeClock.Now(), post.CreatedAt)
	assert.Equal(t, fakeClock.Now(), post.UpdatedAt)
}

func TestLookupComment(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	err := engine.SetPost(t.Context(), &store.Post{
		ID:       "1",
		AuthorID: "someAuthorID",
		Title:    "Some Title",
		Content:  "Some Content",
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       "1",
		AuthorID: "someAuthorID",
		PostID:   "1",
		Content:  "Some Comment",
	})
	require.NoError(t, err)

	comment, err := engine.LookupComment(t.Context(), "1", "1")
	assert.NoError(t, err)
	assert.Equal(t, "1", comment.ID)
	assert.Equal(t, "someAuthorID", comment.AuthorID)
	assert.Equal(t, "1", comment.PostID)
	assert.Equal(t, "Some Comment", comment.Content)
	assert.Equal(t, fakeClock.Now(), comment.CreatedAt)
	assert.Equal(t, fakeClock.Now(), comment.UpdatedAt)

	comment, err = engine.LookupComment(t.Context(), "1", "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, comment)
}

func TestListCommentsByPostID(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	err := engine.SetPost(t.Context(), &store.Post{
		ID:       "1",
		AuthorID: "someAuthorID",
		Title:    "Some Title",
		Content:  "Some Content",
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       "1",
		AuthorID: "someAuthorID",
		PostID:   "1",
		Content:  "First Comment",
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       "2",
		AuthorID: "someAuthorID",
		PostID:   "1",
		Content:  "Second Comment",
	})
	require.NoError(t, err)

	comments, err := engine.ListCommentsByPostID(t.Context(), "1")
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	comments, err = engine.ListCommentsByPostID(t.Context(), "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, comments)
}

func TestDeleteComment(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	err := engine.SetPost(t.Context(), &store.Post{
		ID:       "1",
		AuthorID: "someAuthorID",
		Title:    "Some Title",
		Content:  "Some Content",
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       "1",
		AuthorID: "someAuthorID",
		PostID:   "1",
		Content:  "Some Comment",
	})
	require.NoError(t, err)

	comment, err := engine.LookupComment(t.Context(), "1", "1")
	assert.NoError(t, err)
	assert.NotNil(t, comment)

	err = engine.DeleteComment(t.Context(), "1", "1")
	assert.NoError(t, err)

	comment, err = engine.LookupComment(t.Context(), "1", "1")
	assert.NoError(t, err)
	assert.Nil(t, comment)

	err = engine.DeleteComment(t.Context(), "1", "nonexistent")
	assert.NoError(t, err)
}
