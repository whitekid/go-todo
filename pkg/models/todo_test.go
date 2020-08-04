package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItem(t *testing.T) {
	item := Item{
		ID:      "7dc6140d-de8b-42d8-b845-7fe4ddef3c2e",
		Title:   "title",
		DueDate: Today(),
		Rank:    1,
	}

	buf, err := json.Marshal(&item)
	require.NoError(t, err)
	require.NotEqual(t, "", string(buf))

	var revert Item
	require.NoError(t, json.Unmarshal(buf, &revert))
	require.Equal(t, item, revert, "%s %s", item.DueDate.String(), revert.DueDate.String())
}
