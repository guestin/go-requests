package requests

import (
	"github.com/guestin/go-requests/opt"
	"github.com/guestin/mob/murl"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type GitTag struct {
	Ref    string    `json:"ref"`
	NodeId string    `json:"node_id"`
	TagUrl string    `json:"url"`
	Object SHAObject `json:"object"`
}

type SHAObject struct {
	Sha  string `json:"sha"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

var RemoteGitAPI = NewScope(
	opt.Url("https://api.github.com"))

func TestGetScope(t *testing.T) {
	requestScope := RemoteGitAPI.With(
		opt.HttpMethod(http.MethodGet),
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.BindJSON(&[]GitTag{}))
	data, err := requestScope.Execute()
	assert.NoError(t, err)
	t.Log(data)
}
