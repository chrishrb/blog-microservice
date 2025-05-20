package inmemory_test

import (
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/chrishrb/blog-microservice/post-service/store/inmemory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clock_testing "k8s.io/utils/clock/testing"
)

func TestSetComment(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	ID := uuid.New()
	postID := uuid.New()
	authorID := uuid.New()

	err := engine.SetPost(t.Context(), &store.Post{
		ID:        postID,
		AuthorID:  authorID,
		Title:     "Some Title",
		Content:   "Some Content",
		Tags:      []string{"tag1", "tag2"},
		Published: true,
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       ID,
		AuthorID: authorID,
		PostID:   postID,
		Content:  "Some Comment",
	})
	assert.NoError(t, err)

	comment, err := engine.LookupComment(t.Context(), postID, ID)
	assert.NoError(t, err)
	assert.Equal(t, ID, comment.ID)
	assert.Equal(t, authorID, comment.AuthorID)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, "Some Comment", comment.Content)
	assert.Equal(t, fakeClock.Now(), comment.CreatedAt)
	assert.Equal(t, fakeClock.Now(), comment.UpdatedAt)
}

func TestLookupComment(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	ID := uuid.New()
	postID := uuid.New()
	authorID := uuid.New()

	err := engine.SetPost(t.Context(), &store.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Some Title",
		Content:  "Some Content",
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       ID,
		AuthorID: authorID,
		PostID:   postID,
		Content:  "Some Comment",
	})
	require.NoError(t, err)

	comment, err := engine.LookupComment(t.Context(), postID, ID)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, ID, comment.ID)
	assert.Equal(t, authorID, comment.AuthorID)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, "Some Comment", comment.Content)
	assert.Equal(t, fakeClock.Now(), comment.CreatedAt)
	assert.Equal(t, fakeClock.Now(), comment.UpdatedAt)

	comment, err = engine.LookupComment(t.Context(), postID, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, comment)
}

func TestListCommentsByPostID(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	postID := uuid.New()
	authorID := uuid.New()

	err := engine.SetPost(t.Context(), &store.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Some Title",
		Content:  "Some Content",
	})
	require.NoError(t, err)

	ID1 := uuid.New()
	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       ID1,
		AuthorID: authorID,
		PostID:   postID,
		Content:  "First Comment",
	})
	require.NoError(t, err)

	ID2 := uuid.New()
	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       ID2,
		AuthorID: authorID,
		PostID:   postID,
		Content:  "Second Comment",
	})
	require.NoError(t, err)

	comments, err := engine.ListCommentsByPostID(t.Context(), postID, 0, 100)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	// Test pagination with limit
	limitedDatapoints, err := engine.ListCommentsByPostID(t.Context(), postID, 0, 1)
	assert.NoError(t, err)
	assert.Len(t, limitedDatapoints, 1)

	// Test pagination with offset
	offsetDatapoints, err := engine.ListCommentsByPostID(t.Context(), postID, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, offsetDatapoints, 1)

	// Test nonexistent post
	comments, err = engine.ListCommentsByPostID(t.Context(), uuid.New(), 0, 100)
	assert.NoError(t, err)
	assert.Nil(t, comments)

}

func TestDeleteComment(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	ID := uuid.New()
	postID := uuid.New()
	authorID := uuid.New()

	err := engine.SetPost(t.Context(), &store.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Some Title",
		Content:  "Some Content",
	})
	require.NoError(t, err)

	err = engine.SetComment(t.Context(), &store.Comment{
		ID:       ID,
		AuthorID: authorID,
		PostID:   postID,
		Content:  "Some Comment",
	})
	require.NoError(t, err)

	comment, err := engine.LookupComment(t.Context(), postID, ID)
	assert.NoError(t, err)
	assert.NotNil(t, comment)

	err = engine.DeleteComment(t.Context(), postID, ID)
	assert.NoError(t, err)

	comment, err = engine.LookupComment(t.Context(), postID, ID)
	assert.NoError(t, err)
	assert.Nil(t, comment)

	comment, err = engine.LookupComment(t.Context(), postID, uuid.New())
	assert.NoError(t, err)
}
