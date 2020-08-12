package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO 스토리지 구현체들의 전체 통합 테스트 필요하지요..
func TestStorage(t *testing.T) {
	type args struct {
	}
	tests := [...]struct {
		name string
		args args
	}{
		{"badger", args{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stg, err := New("testdb")
			require.NoError(t, err)
			require.NotNil(t, stg)

			require.NotNil(t, stg.UserService())
			require.NotNil(t, stg.TokenService())
			require.NotNil(t, stg.TodoService())
			// got, err := doSomething()
			// if (err != nil) != tt.wantErr {
			// 	require.Failf(t, `doSomething() failed`, `error = %v, wantErr = %v`, err, tt.wantErr)
			// }
			// _ = got
		})
	}
}
