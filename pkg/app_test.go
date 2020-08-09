package todo

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-todo/pkg/client"
	"github.com/whitekid/go-todo/pkg/models"
	"github.com/whitekid/go-todo/pkg/utils"
)

func newTestServer() (*httptest.Server, string, func()) {
	s := New().(*todoService)
	e := s.setupRoute()

	email := utils.RandomString(5) + "@domain.com"
	t, err := s.storage.TokenService().Create(email)
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(e)
	return ts, t.Token, func() {
		s.storage.UserService().Delete(email)
		ts.Close()
	}
}

func TestTodo(t *testing.T) {
	ts, token, teardown := newTestServer()
	defer teardown()

	item := models.Item{
		Title:   "title",
		DueDate: models.Today(),
		Rank:    1,
	}

	api := client.New(ts.URL, token)

	// create
	var created *models.Item
	{
		item, err := api.Todos.Create(&item)
		require.NoError(t, err)

		created = item
		require.NotEqual(t, "", created.ID)
		item.ID = created.ID
		require.Equal(t, item, created)
	}

	// list
	{
		items, err := api.Todos.List()
		require.NoError(t, err)

		require.Equal(t, 1, len(items), "item created but got %d items", len(items))
		require.Equal(t, []models.Item{*created}, items)
	}

	// get
	{
		item, err := api.Todos.Get(created.ID)
		require.NoError(t, err)

		require.Equal(t, created, item)
	}

	// update
	{
		item := created
		item.Title = "updated title"

		updated, err := api.Todos.Update(item)
		require.NoError(t, err)
		require.Equal(t, updated, item)

		reterived, err := api.Todos.Get(updated.ID)
		require.NoError(t, err)

		require.Equal(t, updated, reterived)
	}

	// delete
	{
		require.NoError(t, api.Todos.Delete(created.ID))

		_, err := api.Todos.Get(created.ID)
		require.Error(t, err)

		items, err := api.Todos.List()
		require.NoError(t, err)
		require.Equal(t, 0, len(items), "item deleted but got %d items", len(items))
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		item models.Item
	}

	tests := [...]struct {
		name     string
		args     args
		wantErr  bool
		wantFail bool
	}{
		{"", args{models.Item{Title: "title", DueDate: models.Today(), Rank: 1}}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, token, teardown := newTestServer()
			defer teardown()
			api := client.New(ts.URL, token)

			created, err := api.Todos.Create(&tt.args.item)
			if (err != nil) != tt.wantErr {
				require.Failf(t, `doSomething() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}

			require.NotEqual(t, "", created.ID)
			tt.args.item.ID = created.ID
			require.Equal(t, &tt.args.item, created)

			got, err := api.Todos.Get(created.ID)
			require.NoError(t, err)
			require.Equal(t, created, got)

			items, err := api.Todos.List()
			require.NoError(t, err)
			require.Equal(t, 1, len(items))
		})
	}
}

func TestList(t *testing.T) {
	type args struct {
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, token, teardown := newTestServer()
			defer teardown()
			api := client.New(ts.URL, token)

			got, err := api.Todos.List()
			if (err != nil) != tt.wantErr {
				require.Failf(t, `List() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}
			_ = got
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		item models.Item
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{models.Item{Title: "title"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, token, teardown := newTestServer()
			defer teardown()
			api := client.New(ts.URL, token)

			created, err := api.Todos.Create(&tt.args.item)
			require.NoError(t, err)

			created.Title = "updated " + created.Title
			got, err := api.Todos.Update(created)
			if (err != nil) != tt.wantErr {
				require.Failf(t, `update() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}
			require.Equal(t, created, got)
		})
	}
}
