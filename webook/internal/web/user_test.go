package web

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestEncrypt(t *testing.T) {
	password := []byte("hello#123")
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(hash))
	err = bcrypt.CompareHashAndPassword(hash, password)
	assert.NoError(t, err)
}
