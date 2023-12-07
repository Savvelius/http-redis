package main

import (
	"bytes"
	"net/http"
	"testing"
)

// TODO
func TestServer(t *testing.T) {
	resp, err := http.Get("http://localhost:8000/reg/sava:1234")
	testStatusCode(t, err, resp.StatusCode, http.StatusOK)
	resp.Body.Close()

	resp, err = http.Post("http://localhost:8000/sava:1234/pairs/key1",
		"application/json", bytes.NewBufferString("\"value1\""))
	testStatusCode(t, err, resp.StatusCode, http.StatusOK)
	resp.Body.Close()

	resp, err = http.Post("http://localhost:8000/sava:1234/pairs/key2",
		"application/json", bytes.NewBufferString("\"value2\""))
	testStatusCode(t, err, resp.StatusCode, http.StatusOK)
	resp.Body.Close()

}

func testStatusCode(t *testing.T, err error, code, targetCode int) {
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if code != targetCode {
		t.Errorf("Expected status to be %d, got %d", targetCode, code)
	}
}
