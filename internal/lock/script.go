package lock

import "github.com/redis/go-redis/v9"

// luaScriptUnlock is a Lua script for atomic lock release.
// It checks if the lock exists and if the token matches before deleting.
var scriptUnlock = `
local val = redis.call("GET", KEYS[1])
if val == false then
    return -1
elseif val == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
`

var luaScriptUnlock = redis.NewScript(scriptUnlock)
