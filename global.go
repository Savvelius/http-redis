package main

import (
	"os"
	"strings"
	"time"
)

type UserInfo struct {
	password string
	active   bool
}

// Invariant: if it was created with NewGlobal()
// then Global.users is filled with all known users.
// Their environments are restored on data access
type Global struct {
	aof   *Aof
	users map[string]UserInfo     // username -> password
	envs  map[string]*Environment // username -> env
}

// Creates new Global environment or restores it from current.
// User environments are restored on their data access
func NewGlobal() Global {
	// create if it doesn't exist
	os.MkdirAll(GLOBAL_AOF_DIR, 0777)

	aof, err := NewAof(GLOBAL_AOF_FILE, time.Second*2)
	if err != nil {
		panic(err)
	}

	bytes, err := aof.ReadAll()
	if err != nil {
		panic(err)
	}

	global := Global{users: map[string]UserInfo{}, envs: map[string]*Environment{}, aof: aof}

	// splitting empty input returns splitter
	if len(bytes) == 0 {
		return global
	}

	// usernames are stored separated by \r\n
	userData := strings.Split(string(bytes), "\r\n")
	for i := 0; i < len(userData)-1; i += 2 {
		username := userData[i]
		password := userData[i+1]

		global.users[username] = UserInfo{password: password, active: false}
	}

	return global
}

// verify user data against user map
func (global *Global) verifyUser(username, password string) bool {
	userInfo, haveUsername := global.users[username]
	return haveUsername && userInfo.password == password
}

// writes data about the new user into the aof
func (global *Global) saveNewUser(username string) {
	global.aof.WriteString(username + "\r\n" + global.users[username].password + "\r\n")
}
