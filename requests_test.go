package requests

import (
	"context"
	"github.com/guestin/go-requests/opt"
	"github.com/guestin/mob/murl"
	"github.com/guestin/mob/mvalidate"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	arr := &[]struct {
		Ref string `json:"ref"`
		Url string `json:"url"`
	}{}
	data, err := Get(context.TODO(),
		"https://api.github.com",
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.ExpectStatusCode(http.StatusOK),
		opt.ResponseBodyDump(os.Stdout),
		opt.BindJSON(arr))
	assert.NoError(t, err)
	t.Log(data)
}

func TestValidateResponse(t *testing.T) {
	validator, err := mvalidate.NewValidator("zh")
	assert.NoError(t, err)
	data, err := Get(context.TODO(),
		"https://api.github.com",
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.ExpectStatusCode(http.StatusOK),
		opt.BindJSON(&[]GitTag{}),
		opt.ResponseBodyDump(os.Stdout),
		opt.ValidateVar(validator, `required,dive,required`))
	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}
	versionData := data.(*[]GitTag)
	t.Logf("%v\n", versionData)
}
