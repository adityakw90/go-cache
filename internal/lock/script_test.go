package lock

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestLock_LuaScriptUnlock(t *testing.T) {
	client, mock := redismock.NewClientMock()
	key := "test:lock:1"
	token := "1234567890"
	// Mock script execution: EvalSha tries first, but scripts aren't cached in tests
	mock.Regexp().ExpectEvalSha(`.*`, []string{key}, []interface{}{token}).SetErr(errors.New("NOSCRIPT "))
	mock.ExpectEval(scriptUnlock, []string{key}, []interface{}{token}).SetVal(int64(1))
	result, err := luaScriptUnlock.Run(context.Background(), client, []string{key}, token).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
