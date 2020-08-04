package todo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
		DueDate: Today(),
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

		resp, err := sess.Put("%s/%s", ts.URL, updated.ID).JSON(&updated).Do()
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

func TestCreate(t *testing.T) {
	type args struct {
		item todoItem
	}

	tests := [...]struct {
		name     string
		args     args
		wantErr  bool
		wantFail bool
	}{
		{"", args{todoItem{Title: "title", DueDate: Today(), Rank: 1}}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, teardown := newTestServer()
			defer teardown()

			sess := request.NewSession(nil)

			item := tt.args.item

			resp, err := sess.Post("%s/", ts.URL).JSON(&item).Do()
			if (err != nil) != tt.wantErr {
				require.Failf(t, `doSomething() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}

			if (!resp.Success()) != tt.wantFail {
				require.Failf(t, "creaet failed", "status = %d", resp.StatusCode)
			}

			var created todoItem
			defer resp.Body.Close()
			require.NoError(t, resp.JSON(&created))
			require.NotEqual(t, "", created.ID)
			item.ID = created.ID
			require.Equal(t, item, created)

			cookies := resp.Cookies()
			require.NotEqual(t, 0, len(cookies), "need to be set cookie")
		})
	}
}

func TestTodoItem(t *testing.T) {
	item := todoItem{
		ID:      "7dc6140d-de8b-42d8-b845-7fe4ddef3c2e",
		Title:   "title",
		DueDate: Today(),
		Rank:    1,
	}

	buf, err := json.Marshal(&item)
	require.NoError(t, err)
	require.NotEqual(t, "", string(buf))

	var revert todoItem
	require.NoError(t, json.Unmarshal(buf, &revert))
	require.Equal(t, item, revert, "%s %s", item.DueDate.String(), revert.DueDate.String())
}

func doSomething() (string, error) { return "something", nil }
