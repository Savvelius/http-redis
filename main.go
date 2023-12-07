package main

import "fmt"

const PORT = ":8000"

func main() {
	srv := NewServer(PORT)
	fmt.Printf("Listening on port %s", PORT)
	srv.ListenAndServe()
}
