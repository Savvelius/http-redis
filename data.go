package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrHashDoesntExist = errors.New("hash of given key doesn't exist")
)

type Environment struct {
	aof   *Aof
	Pairs map[string]string
	Hash  map[string]map[string]string
}

func NewEnv(username string) *Environment {
	aof, err := NewAof(getUserAofPath(username), time.Second)
	if err != nil {
		panic(err)
	}

	if aof.IsEmpty() {
		aof.Write([]byte("[]"))
	}

	return &Environment{
		Pairs: map[string]string{},
		Hash:  map[string]map[string]string{},
		aof:   aof,
	}
}

func (env *Environment) DeletePair(key string) {
	delete(env.Pairs, key)
}

func (env *Environment) DeleteAllPairs() {
	for k := range env.Pairs {
		delete(env.Pairs, k)
	}
}

func (env *Environment) GetPair(key string) string {
	return env.Pairs[key]
}

func (env *Environment) GetAllPairs() map[string]string {
	return env.Pairs
}

func (env *Environment) SetPair(key, val string) {
	env.aof.WriteCommand("SetPair", key, val)
	env.Pairs[key] = val
}

func (env *Environment) DeleteHash(key string) {
	delete(env.Hash, key)
}

func (env *Environment) DeleteHashVal(key1, key2 string) bool {
	if _, ok := env.Hash[key1]; !ok {
		return false
	}
	delete(env.Hash[key1], key2)
	return true
}

func (env *Environment) DeleteAllHash() {
	for k := range env.Hash {
		delete(env.Hash, k)
	}
}

func (env *Environment) GetAllHash(key string) map[string]string {
	return env.Hash[key]
}

func (env *Environment) GetKeyHash(key1, key2 string) (string, error) {
	queriedMap, ok := env.Hash[key1]
	if !ok {
		return "", ErrHashDoesntExist
	}

	return queriedMap[key2], nil
}

func (env *Environment) SetHash(key string, pairs map[string]string) {
	env.aof.WriteCommand("SetHash", key, pairs)

	if _, ok := env.Hash[key]; !ok {
		env.Hash[key] = map[string]string{}
	}

	for k, v := range pairs {
		env.Hash[key][k] = v
	}
}

type Command struct {
	FuncName string `json:"funcName"`
	Args     []any  `json:"args"`
}

func JsonToReader(val any) (io.Reader, error) {
	encoded, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(encoded), nil
}

func (env *Environment) Restore() error {
	commands, err := env.aof.ReadCommands()
	if err != nil {
		return err
	}

	for _, cmd := range commands {
		switch cmd.FuncName {
		case "SetPair":
			env.SetPair(cmd.Args[0].(string), cmd.Args[1].(string))
		case "SetHash":
			env.SetHash(cmd.Args[0].(string), cmd.Args[1].(map[string]string))
		default:
			return fmt.Errorf("illegal function name (%s) was written into Aof", cmd.FuncName)
		}
	}

	return nil
}
