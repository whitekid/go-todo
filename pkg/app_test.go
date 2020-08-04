package todo

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-utils/request"
)

func newTestServer() (*httptest.Server, func()) {
	s := New().(*todoService)
	e := s.setupRoute()

	ts := httptest.NewServer(e)
	return ts, func() { ts.Close() }
}

func TestTodo(t *testing.T) {
	ts, teardown := newTestServer()
	defer teardown()

	item := todoItem{
		Title:   "title",
		DueDate: time.Now().Truncate(0),
		Rank:    1,
	}

	sess := request.NewSession(nil)

	// create
	var created todoItem
	{
		resp, err := sess.Post("%s/", ts.URL).JSON(&item).Do()
		require.NoError(t, err)
		require.True(t, resp.Success(), "create failed with %d", resp.StatusCode)

		defer resp.Body.Close()
		require.NoError(t, resp.JSON(&created))
		require.NotEqual(t, "", created.ID)
		item.ID = created.ID
		require.Equal(t, item, created)

		cookies := resp.Cookies()
		require.NotEqual(t, 0, len(cookies), "need to be set cookie")
	}

	// list
	{
		resp, err := sess.Get("%s", ts.URL).Do()
		require.NoError(t, err)
		require.True(t, resp.Success(), "failed with status %d", resp.StatusCode)

		items := make([]todoItem, 0)
		defer resp.Body.Close()
		require.NoError(t, resp.JSON(&items))
		require.Equal(t, 1, len(items), "item created but got %d items", len(items))
		require.Equal(t, []todoItem{created}, items)
	}

	// get
	{
		resp, err := sess.Get("%s/%s", ts.URL, created.ID).Do()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var item todoItem
		defer resp.Body.Close()
		require.NoError(t, resp.JSON(&item))

		require.Equal(t, created, item)
	}

	// update
	{
		updated := created
		updated.Title = "updated title"

		resp, err := sess.Put("%s/%s", ts.URL, updated.ID).JSON(updated).Do()
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var item todoItem
		defer resp.Body.Close()
		require.NoError(t, resp.JSON(&item))

		require.Equal(t, updated, item)

		resp, err = sess.Get("%s/%s", ts.URL, updated.ID).Do()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var reterived todoItem
		defer resp.Body.Close()
		require.NoError(t, resp.JSON(&reterived))
		require.Equal(t, updated, reterived)
	}

	// delete
	{
		resp, err := sess.Delete("%s/%s", ts.URL, created.ID).Do()
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		resp, err = sess.Get("%s/%s", ts.URL, created.ID).Do()
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		resp, err = sess.Get("%s", ts.URL).Do()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		items := make([]todoItem, 0)
		defer resp.Body.Close()
		require.NoError(t, resp.JSON(&items))
		require.Equal(t, 0, len(items), "item deleted but got %d items", len(items))
	}
}
