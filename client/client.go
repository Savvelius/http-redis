package client

import (
	"fmt"
	"net/http"
)

type Connection struct {
	AuthString string
	BindAddr   string
	Client     *http.Client
}

const CONTENT_TYPE = "application/json"

func NewConnection(username, password, bindAddr string) (*Connection, error) {
	conn := &Connection{
		AuthString: username + ":" + password,
		BindAddr:   bindAddr,
		Client:     &http.Client{},
	}
	resp, err := conn.Client.Post(bindAddr+"/reg/"+conn.AuthString, CONTENT_TYPE, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to connect: responsed with status: %d", resp.StatusCode)
	}
	return conn, nil
}
