package server

import (
	"errors"
	"net/http"
)

type Method string

type Node struct {
	Next       []Node                      // next paths
	Pattern    string                      // on what path does it trigger
	Handlers   map[Method]http.HandlerFunc // method -> handler
	Middleware []http.HandlerFunc
}

func NewRouter() Node {
	return Node{
		Next:    []Node{},
		Pattern: "/",
	}
}

func (node *Node) addHahdler(method Method, handler http.HandlerFunc) {
	if node.Handlers == nil {
		node.Handlers = map[Method]http.HandlerFunc{}
	}
	node.Handlers[method] = handler
}

func (node *Node) AddMiddleware(middleware http.HandlerFunc) {
	if node.Middleware == nil {
		node.Middleware = []http.HandlerFunc{middleware}
	} else {
		node.Middleware = append(node.Middleware, middleware)
	}
}

var ErrBadPath = errors.New("cannot route given path")

func (node *Node) routePath(path string, walker func(*Node)) *Node {
	step := len(node.Pattern)
	if step > len(path) || node.Pattern != path[:step] {
		return nil
	}

	path = path[:step]
	if len(path) == 0 {
		return node
	}

	if walker != nil {
		walker(node)
	}

	for _, next := range node.Next {
		if target := next.routePath(path, walker); target != nil {
			return target
		}
	}

	return nil
}

func (node *Node) ConnectHandler(method Method, path string, handler http.HandlerFunc) error {
	if found := node.routePath(path, nil); found != nil {
		found.addHahdler(method, handler)
		return nil
	}
	return ErrBadPath
}

func (node *Node) Get(path string, handler http.HandlerFunc) error {
	return node.ConnectHandler("GET", path, handler)
}

func (node *Node) Post(path string, handler http.HandlerFunc) error {
	return node.ConnectHandler("POST", path, handler)
}

func (node *Node) ConnectMiddleware(path string, middleware http.HandlerFunc) error {
	if found := node.routePath(path, nil); found != nil {
		found.AddMiddleware(middleware)
		return nil
	}
	return ErrBadPath
}

func (node *Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	endPoint := node.routePath(path, func(n *Node) {
		for _, middleware := range n.Middleware {
			middleware(w, r)
		}
	})

	if endPoint == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	method := r.Method
	if method == "" {
		method = "GET"
	}

	handler, ok := endPoint.Handlers[Method(method)]
	if !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	handler(w, r)
}
