package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartServer(t *testing.T) {
	s, _, _, err := StartServer()
	// if err = s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
	// 	e.Logger.Fatal(err)
	// }
	if assert.NoError(t, err) {
		s.Close()
	}
}
