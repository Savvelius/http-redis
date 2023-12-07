package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const (
	GLOBAL_AOF_DIR  = "internal/logs"
	GLOBAL_AOF_FILE = "internal/global_log.log"
)

// key for retrieving user's environment from context
var CTX_KEY = struct{}{}

type apiError struct {
	Err    string `json:"error"`
	Status int    `json:"status"`
}

func (err apiError) Error() string {
	return fmt.Sprintf("%s, status: %d", err.Err, err.Status)
}

var (
	ErrAuthRequired      = apiError{Err: "activity requires authorization", Status: http.StatusUnauthorized}
	ErrUserAlreadyExists = apiError{Err: "user with such username already exists", Status: http.StatusBadRequest}
	ErrDecode            = apiError{Err: "unable to decode your command", Status: http.StatusUnprocessableEntity}
	ErrObjectDoesntExist = apiError{Err: "queried object doesn't exist", Status: http.StatusNotFound}
)

type Server struct {
	http.Server
}

func NewServer(port string) *Server {
	r := chi.NewRouter()
	global := NewGlobal()

	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentType("application/json"))

	r.Post("/reg/{username}:{password}", global.RegisterUser)
	r.Post("/quit/{username}:{password}", global.DeleteSession)
	r.Delete("/clear/{username}:{password}", global.ClearUserData)

	v1 := chi.NewRouter()
	r.Mount("/{username}:{password}", v1)

	v1.Use(global.AuthorizationMiddleware)

	v1.Get("/pairs/{key}", GetPair)
	v1.Get("/pairs", GetAllPairs)
	v1.Delete("/pairs/{key}", DeletePair)
	v1.Delete("/pairs", DeleteAllPairs)
	v1.Post("/pairs/{key}", PostPair)

	v1.Get("/hash/{key1}/{key2}", GetHash)
	v1.Get("/hash/{key}", GetAllHash)
	v1.Delete("/hash/{key}", DeleteHash)
	v1.Delete("/hash/{key1}/{key2}", DeleteHashVal)
	v1.Delete("/hash", DeleteAllHash)
	v1.Post("/hash/{key}", PostHash)

	return &Server{
		Server: http.Server{
			Addr:    port,
			Handler: r,
		},
	}
}

func getUserAofPath(username string) string {
	return path.Join(GLOBAL_AOF_DIR, username+".json")
}

func extractEnv(r *http.Request) *Environment {
	return r.Context().Value(CTX_KEY).(*Environment)
}

func writeJson(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	bytes, _ := json.Marshal(value)
	w.Write(bytes)
}

func DeletePair(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	env := extractEnv(r)

	env.DeletePair(key)
	w.WriteHeader(http.StatusOK)
}

func DeleteAllPairs(w http.ResponseWriter, r *http.Request) {
	env := extractEnv(r)

	env.DeleteAllPairs()
	w.WriteHeader(http.StatusOK)
}

func PostPair(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	env := extractEnv(r)

	var val string
	err := json.NewDecoder(r.Body).Decode(&val)

	if err != nil {
		writeJson(w, http.StatusUnprocessableEntity, ErrDecode)
		return
	}

	env.SetPair(key, val)
	w.WriteHeader(http.StatusOK)
}

func GetPair(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	env := extractEnv(r)

	value := env.GetPair(key)
	if value == "" {
		writeJson(w, http.StatusNotFound, ErrObjectDoesntExist)
		return
	}

	writeJson(w, http.StatusOK, value)
}

func GetAllPairs(w http.ResponseWriter, r *http.Request) {
	env := extractEnv(r)

	queried := env.GetAllPairs()
	writeJson(w, http.StatusOK, queried)
}

func DeleteHash(w http.ResponseWriter, r *http.Request) {
	env := extractEnv(r)
	key := chi.URLParam(r, "key")

	env.DeleteHash(key)
	w.WriteHeader(http.StatusOK)
}

func DeleteHashVal(w http.ResponseWriter, r *http.Request) {
	env := extractEnv(r)
	key1 := chi.URLParam(r, "key1")
	key2 := chi.URLParam(r, "key2")

	if !env.DeleteHashVal(key1, key2) {
		writeJson(w, http.StatusNotFound, ErrObjectDoesntExist)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteAllHash(w http.ResponseWriter, r *http.Request) {
	env := extractEnv(r)
	env.DeleteAllHash()
}

func GetAllHash(w http.ResponseWriter, r *http.Request) {
	var (
		key     = chi.URLParam(r, "key")
		env     = extractEnv(r)
		queried = env.GetAllHash(key)
	)
	writeJson(w, http.StatusOK, queried)
}

func PostHash(w http.ResponseWriter, r *http.Request) {
	var (
		key      = chi.URLParam(r, "key")
		env      = extractEnv(r)
		commands = map[string]string{}
		err      = json.NewDecoder(r.Body).Decode(&commands)
	)

	if err != nil {
		writeJson(w, http.StatusUnprocessableEntity, ErrDecode)
		return
	}

	env.SetHash(key, commands)
	w.WriteHeader(http.StatusOK)
}

func GetHash(w http.ResponseWriter, r *http.Request) {
	key1 := chi.URLParam(r, "key1")
	key2 := chi.URLParam(r, "key2")
	env := extractEnv(r)

	val, err := env.GetKeyHash(key1, key2)

	if err != nil {
		writeJson(w, http.StatusNotFound, apiError{Err: err.Error(), Status: http.StatusNotFound})
		return
	}

	if val == "" {
		writeJson(w, http.StatusNotFound, ErrObjectDoesntExist)
		return
	}

	writeJson(w, http.StatusOK, val)
}

func (global *Global) ClearUserData(w http.ResponseWriter, r *http.Request) {
	var (
		username = chi.URLParam(r, "username")
		password = chi.URLParam(r, "password")
	)

	if !global.verifyUser(username, password) {
		writeJson(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	env := extractEnv(r)
	env.DeleteAllHash()
	env.DeleteAllPairs()
	if err := env.aof.Clear(); err != nil {
		writeJson(w, http.StatusInternalServerError,
			apiError{Err: "unable to clear your data", Status: http.StatusInternalServerError})
		fmt.Printf("Error clearing data of user %s, err: %v\n", username, err)
		return
	}
}

func (global *Global) DeleteSession(w http.ResponseWriter, r *http.Request) {
	var (
		username = chi.URLParam(r, "username")
		password = chi.URLParam(r, "password")
	)

	if !global.verifyUser(username, password) {
		writeJson(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	userInfo := global.users[username]
	if !userInfo.active {
		writeJson(w, http.StatusBadRequest, apiError{Err: "you are already logged out", Status: http.StatusBadRequest})
		return
	}

	env := extractEnv(r)
	env.aof.Drop()
	delete(global.envs, username)

	userInfo.active = false
	global.users[username] = userInfo

	w.WriteHeader(http.StatusOK)
}

func (global *Global) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var (
		username = chi.URLParam(r, "username")
		password = chi.URLParam(r, "password")
	)

	if userInfo, haveUsername := global.users[username]; haveUsername {
		if userInfo.password == password {
			writeJson(w, http.StatusBadRequest, apiError{
				Err:    "you are already registered",
				Status: http.StatusBadRequest,
			})
			return
		}
		writeJson(w, http.StatusBadRequest, ErrUserAlreadyExists)
		return
	}

	global.users[username] = UserInfo{password: password, active: true}
	if _, ok := global.envs[username]; !ok {
		global.envs[username] = NewEnv(username)
	}
	global.saveNewUser(username)
	w.WriteHeader(http.StatusOK)
}

func (global *Global) AuthorizationMiddleware(initial http.Handler) http.Handler {
	hf := func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		password := chi.URLParam(r, "password")

		if !global.verifyUser(username, password) {
			writeJson(w, http.StatusUnauthorized, ErrAuthRequired)
			return
		}

		if _, haveEnv := global.envs[username]; !haveEnv {
			global.envs[username] = NewEnv(username)
			err := global.envs[username].Restore()
			if err != nil {
				fmt.Printf("Error restoring user data: %v\n", err)
			}

			userInfo := global.users[username]
			if userInfo.active {
				panic("user should be inactive")
			}
			userInfo.active = true
			global.users[username] = userInfo
		}

		initial.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CTX_KEY, global.envs[username])))
	}
	return http.HandlerFunc(hf)
}
