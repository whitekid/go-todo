package badger

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	. "github.com/whitekid/go-todo/storage/types"
	"github.com/whitekid/go-utils/fixtures"
)

func storageFixture(callbacks ...func(Interface)) func() {
	var dir string
	defer fixtures.TempDir("testdb", "test_", func(tempDir string) { dir = tempDir })()

	s, err := New(dir)
	if err != nil {
		panic(err)
	}

	for _, callback := range callbacks {
		callback(s)
	}

	return func() {
		defer s.Close()
	}
}

func TestBadger(t *testing.T) {
	var dir string
	defer fixtures.TempDir(".", "testdb_", func(tempDir string) { dir = tempDir })()

	s, err := New(dir)
	defer s.Close()
	require.NoError(t, err)

	todos := s.TodoService()
	tokens := s.TokenService()

	email := "whitekid@gmail.com"

	refreshToken, err := tokens.Create(email)
	require.NoError(t, err)
	require.Equal(t, email, refreshToken.Email)

	{
		got, err := tokens.Get(refreshToken.Token)
		require.NoError(t, err)
		require.Equal(t, got.Email, email)
	}

	item := TodoItem{
		ID:    uuid.New().String(),
		Title: "title",
	}
	require.NoError(t, todos.Create(email, &item))

	items, err := todos.List(email)
	require.NoError(t, err)
	require.Equal(t, []TodoItem{item}, items)

	got, err := todos.Get(email, item.ID)
	require.NoError(t, err)
	require.Equal(t, &item, got)

	got.Title = "updated"
	require.NoError(t, todos.Update(email, got))

	require.NoError(t, todos.Delete(email, got.ID))
	if _, err := todos.Get(email, item.ID); err == nil {
		require.Fail(t, "should error", "want = %v, got = %v", ErrNotFound, err)
	}

	{
		items, err := todos.List(email)
		require.NoError(t, err)
		require.Equal(t, 0, len(items))
	}
}
