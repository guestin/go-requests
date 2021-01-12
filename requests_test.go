package requests

import (
	"context"
	"github.com/guestin/go-requests/opt"
	"github.com/guestin/mob/murl"
	"github.com/guestin/mob/mvalidate"

	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGet(t *testing.T) {
	data, err := Get(context.TODO(),
		"https://api.github.com",
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.ExpectStatusCode(http.StatusOK))
	assert.NoError(t, err)
	t.Log(data)
}

func TestValidateResponse(t *testing.T) {
	validator, err := mvalidate.NewValidator("zh")
	assert.NoError(t, err)
	data, err := Get(context.TODO(),
		"https://api.github.com",
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.ValidateVar(validator, `required,dive,required`),
		opt.ExpectStatusCode(http.StatusOK),
		opt.BindJSON(&[]GitTag{}))
	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}
	versionData := data.(*[]GitTag)
	t.Logf("%v\n", versionData)
}
