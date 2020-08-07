package badger

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	. "github.com/whitekid/go-todo/pkg/storage/types"
	"github.com/whitekid/go-utils/fixtures"
)

func TestBadger(t *testing.T) {
	var dir string
	defer fixtures.TempDir(".", "testdb_", func(tempDir string) { dir = tempDir })()

	s, err := New(dir)
	defer s.Close()
	require.NoError(t, err)

	todos := s.TodoService()

	item := TodoItem{
		ID:    uuid.New().String(),
		Title: "title",
	}
	require.NoError(t, todos.Create(&item))

	items, err := todos.List()
	require.NoError(t, err)
	require.Equal(t, []TodoItem{item}, items)

	got, err := todos.Get(item.ID)
	require.NoError(t, err)
	require.Equal(t, &item, got)

	got.Title = "updated"
	require.NoError(t, todos.Update(got))

	require.NoError(t, todos.Delete(got.ID))
	if _, err := todos.Get(item.ID); err == nil {
		require.Fail(t, "should error", "want = %v, got = %v", ErrNotFound, err)
	}

	{
		items, err := todos.List()
		require.NoError(t, err)
		require.Equal(t, 0, len(items))
	}
}
