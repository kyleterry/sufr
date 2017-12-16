package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPITokenCreateAndGet(t *testing.T) {
	token, err := CreateAPIToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token.Token)
	assert.NotNil(t, token.ID)

	tkuuid := token.Token

	{
		token, err := GetAPIToken()
		assert.NoError(t, err)

		assert.Equal(t, token.Token, tkuuid)
	}
}

func TestDeleteAPITokens(t *testing.T) {
	token, err := CreateAPIToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token.Token)
	assert.NotNil(t, token.ID)

	assert.NoError(t, DeleteAPITokens())

	{
		token, err := GetAPIToken()
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, token)
	}
}
