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

func TestSetPost(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	ID := uuid.New()
	authorID := uuid.New()

	err := engine.SetPost(t.Context(), &store.Post{
		ID:        ID,
		AuthorID:  authorID,
		Title:     "Some Title",
		Content:   "Some Content",
		Tags:      []string{"tag1", "tag2"},
		Published: true,
	})
	require.NoError(t, err)

	post, err := engine.LookupPost(t.Context(), ID)
	require.NoError(t, err)
	assert.Equal(t, ID, post.ID)
	assert.Equal(t, authorID, post.AuthorID)
	assert.Equal(t, "Some Title", post.Title)
	assert.Equal(t, "Some Content", post.Content)
	assert.Equal(t, []string{"tag1", "tag2"}, post.Tags)
	assert.True(t, post.Published)
	assert.Equal(t, fakeClock.Now(), post.CreatedAt)
	assert.Equal(t, fakeClock.Now(), post.UpdatedAt)
}

func TestLookupPost(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	ID := uuid.New()
	authorID := uuid.New()

	post := &store.Post{
		ID:        ID,
		AuthorID:  authorID,
		Title:     "Some Title",
		Content:   "Some Content",
		Tags:      []string{"tag1", "tag2"},
		Published: true,
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	result, err := engine.LookupPost(t.Context(), ID)
	require.NoError(t, err)
	assert.Equal(t, post.ID, result.ID)
	assert.Equal(t, post.AuthorID, result.AuthorID)
	assert.Equal(t, post.Title, result.Title)
	assert.Equal(t, post.Content, result.Content)
	assert.Equal(t, post.Tags, result.Tags)
	assert.Equal(t, post.Published, result.Published)
	assert.Equal(t, fakeClock.Now(), result.CreatedAt)
	assert.Equal(t, fakeClock.Now(), result.UpdatedAt)

	result, err = engine.LookupPost(t.Context(), uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestListPosts(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	ID1 := uuid.New()
	authorID1 := uuid.New()
	err := engine.SetPost(t.Context(), &store.Post{
		ID:        ID1,
		AuthorID:  authorID1,
		Title:     "Title 1",
		Content:   "Content 1",
		Tags:      []string{"tag1"},
		Published: true,
	})
	require.NoError(t, err)

	ID2 := uuid.New()
	authorID2 := uuid.New()
	err = engine.SetPost(t.Context(), &store.Post{
		ID:        ID2,
		AuthorID:  authorID2,
		Title:     "Title 2",
		Content:   "Content 2",
		Tags:      []string{"tag2"},
		Published: false,
	})
	require.NoError(t, err)

	posts, err := engine.ListPosts(t.Context(), 0, 100)
	require.NoError(t, err)
	assert.Len(t, posts, 2)

	ids := map[uuid.UUID]bool{
		posts[0].ID: true,
		posts[1].ID: true,
	}
	assert.True(t, ids[ID1])
	assert.True(t, ids[ID2])

	// Test pagination with limit
	limitedDatapoints, err := engine.ListPosts(t.Context(), 0, 1)
	assert.NoError(t, err)
	assert.Len(t, limitedDatapoints, 1)

	// Test pagination with offset
	offsetDatapoints, err := engine.ListPosts(t.Context(), 1, 1)
	assert.NoError(t, err)
	assert.Len(t, offsetDatapoints, 1)
}

func TestDeletePost(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	ID := uuid.New()
	authorID := uuid.New()

	post := &store.Post{
		ID:       ID,
		AuthorID: authorID,
		Title:    "Some Title",
		Content:  "Some Content",
	}
	err := engine.SetPost(t.Context(), post)
	require.NoError(t, err)

	_, err = engine.LookupPost(t.Context(), ID)
	require.NoError(t, err)

	err = engine.DeletePost(t.Context(), ID)
	require.NoError(t, err)

	result, err := engine.LookupPost(t.Context(), ID)
	assert.NoError(t, err)
	assert.Nil(t, result)

	err = engine.DeletePost(t.Context(), uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, result)
}
