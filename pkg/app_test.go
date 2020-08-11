package todo

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-todo/pkg/client"
	"github.com/whitekid/go-todo/pkg/config"
	"github.com/whitekid/go-todo/pkg/models"
	"github.com/whitekid/go-todo/pkg/tokens"
	"github.com/whitekid/go-utils"
)

func newTestServer(t *testing.T) (*httptest.Server, string, func()) {
	s := New().(*todoService)
	e := s.setupRoute()

	email := utils.RandomString(5) + "@domain.com"
	refreshToken, err := tokens.New(email, config.RefreshTokenDuration())
	require.NoError(t, err)
	require.NoError(t, s.storage.TokenService().Create(email, refreshToken))
	accessToken, err := tokens.New(email, config.AccessTokenDuration())
	require.NoError(t, err)

	ts := httptest.NewServer(e)
	return ts, accessToken, func() {
		s.storage.UserService().Delete(email)
		ts.Close()
	}
}

func TestTodo(t *testing.T) {
	ts, token, teardown := newTestServer(t)
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
		item, err := api.TodoService().Create(&item)
		require.NoError(t, err)

		created = item
		require.NotEqual(t, "", created.ID)
		item.ID = created.ID
		require.Equal(t, item, created)
	}

	// list
	{
		items, err := api.TodoService().List()
		require.NoError(t, err)

		require.Equal(t, 1, len(items), "item created but got %d items", len(items))
		require.Equal(t, []models.Item{*created}, items)
	}

	// get
	{
		item, err := api.TodoService().Get(created.ID)
		require.NoError(t, err)

		require.Equal(t, created, item)
	}

	// update
	{
		item := created
		item.Title = "updated title"

		updated, err := api.TodoService().Update(item)
		require.NoError(t, err)
		require.Equal(t, updated, item)

		reterived, err := api.TodoService().Get(updated.ID)
		require.NoError(t, err)

		require.Equal(t, updated, reterived)
	}

	// delete
	{
		require.NoError(t, api.TodoService().Delete(created.ID))

		_, err := api.TodoService().Get(created.ID)
		require.Error(t, err)

		items, err := api.TodoService().List()
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
			ts, token, teardown := newTestServer(t)
			defer teardown()
			api := client.New(ts.URL, token)

			created, err := api.TodoService().Create(&tt.args.item)
			if (err != nil) != tt.wantErr {
				require.Failf(t, `create() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}

			require.NotEqual(t, "", created.ID)
			tt.args.item.ID = created.ID
			require.Equal(t, &tt.args.item, created)

			got, err := api.TodoService().Get(created.ID)
			require.NoError(t, err)
			require.Equal(t, created, got)

			items, err := api.TodoService().List()
			require.NoError(t, err)
			require.Equal(t, 1, len(items))

			require.NoError(t, api.TodoService().Delete(created.ID))
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
			ts, token, teardown := newTestServer(t)
			defer teardown()
			api := client.New(ts.URL, token)

			got, err := api.TodoService().List()
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
			ts, token, teardown := newTestServer(t)
			defer teardown()
			api := client.New(ts.URL, token)

			created, err := api.TodoService().Create(&tt.args.item)
			require.NoError(t, err)

			created.Title = "updated " + created.Title
			got, err := api.TodoService().Update(created)
			if (err != nil) != tt.wantErr {
				require.Failf(t, `update() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			}
			require.Equal(t, created, got)
		})
	}
}
