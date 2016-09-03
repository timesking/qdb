// Copyright 2015 Reborndb Org. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package service

import (
	redis "github.com/reborndb/go/redis/resp"
	"github.com/timesking/qdb/pkg/store"
)

// HGETALL key
func HGetAllCmd(s Session, args [][]byte) (redis.Resp, error) {
	if a, err := s.Store().HGetAll(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		resp := redis.NewArray()
		for _, v := range a {
			resp.AppendBulkBytes(v)
		}
		return resp, nil
	}
}

// HDEL key field [field ...]
func HDelCmd(s Session, args [][]byte) (redis.Resp, error) {
	if n, err := s.Store().HDel(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewInt(n), nil
	}
}

// HEXISTS key field
func HExistsCmd(s Session, args [][]byte) (redis.Resp, error) {
	if x, err := s.Store().HExists(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewInt(x), nil
	}
}

// HGET key field
func HGetCmd(s Session, args [][]byte) (redis.Resp, error) {
	if b, err := s.Store().HGet(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewBulkBytes(b), nil
	}
}

// HLEN key
func HLenCmd(s Session, args [][]byte) (redis.Resp, error) {
	if n, err := s.Store().HLen(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewInt(n), nil
	}
}

// HINCRBY key field delta
func HIncrByCmd(s Session, args [][]byte) (redis.Resp, error) {
	if v, err := s.Store().HIncrBy(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewInt(v), nil
	}
}

// HINCRBYFLOAT key field delta
func HIncrByFloatCmd(s Session, args [][]byte) (redis.Resp, error) {
	if v, err := s.Store().HIncrByFloat(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewBulkBytesWithString(store.FormatFloatString(v)), nil
	}
}

// HKEYS key
func HKeysCmd(s Session, args [][]byte) (redis.Resp, error) {
	if a, err := s.Store().HKeys(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		resp := redis.NewArray()
		for _, v := range a {
			resp.AppendBulkBytes(v)
		}
		return resp, nil
	}
}

// HVALS key
func HValsCmd(s Session, args [][]byte) (redis.Resp, error) {
	if a, err := s.Store().HVals(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		resp := redis.NewArray()
		for _, v := range a {
			resp.AppendBulkBytes(v)
		}
		return resp, nil
	}
}

// HSET key field value
func HSetCmd(s Session, args [][]byte) (redis.Resp, error) {
	if x, err := s.Store().HSet(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewInt(x), nil
	}
}

// HSETNX key field value
func HSetNXCmd(s Session, args [][]byte) (redis.Resp, error) {
	if x, err := s.Store().HSetNX(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewInt(x), nil
	}
}

// HMSET key field value [field value ...]
func HMSetCmd(s Session, args [][]byte) (redis.Resp, error) {
	if err := s.Store().HMSet(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		return redis.NewString("OK"), nil
	}
}

// HMGET key field [field ...]
func HMGetCmd(s Session, args [][]byte) (redis.Resp, error) {
	if a, err := s.Store().HMGet(s.DB(), args); err != nil {
		return toRespError(err)
	} else {
		resp := redis.NewArray()
		for _, v := range a {
			resp.AppendBulkBytes(v)
		}
		return resp, nil
	}
}

func init() {
	Register("hdel", HDelCmd, CmdWrite)
	Register("hexists", HExistsCmd, CmdReadonly)
	Register("hget", HGetCmd, CmdReadonly)
	Register("hgetall", HGetAllCmd, CmdReadonly)
	Register("hincrby", HIncrByCmd, CmdWrite)
	Register("hincrbyfloat", HIncrByFloatCmd, CmdWrite)
	Register("hkeys", HKeysCmd, CmdReadonly)
	Register("hlen", HLenCmd, CmdReadonly)
	Register("hmget", HMGetCmd, CmdReadonly)
	Register("hmset", HMSetCmd, CmdWrite)
	Register("hset", HSetCmd, CmdWrite)
	Register("hsetnx", HSetNXCmd, CmdWrite)
	Register("hvals", HValsCmd, CmdReadonly)
}
