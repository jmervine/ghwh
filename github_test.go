package github

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

var (
	payload = `{
        "ref": "ref/head/test-branch",
        "repository": {
            "full_name": "heroku/access-requests"
        },
        "head_commit": {
            "modified": [
                "splunk-users-test.yml"
            ]
        }
    }`
)

func init() {
	Init(logrus.New().WithField("test", true))
}

func TestGithub_InvalidBranchError(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(InvalidBranchError().Error(), "Github webhook Ref contains invalid branch, expected test-branch")
}

func TestGithub_Fetch(t *testing.T) {
	assert := assert.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.github.com/repos/heroku/access-requests/contents/splunk-users-test.yml?ref=test-branch",
		httpmock.NewStringResponder(200, payload))

	body, err := Fetch("splunk-users-test.yml")

	assert.Nil(err)
	assert.Equal(string(body), payload)
}

func TestGithub_Decode(t *testing.T) {
	assert := assert.New(t)
	webhook := new(WebhookPayload)
	reader := strings.NewReader(payload)
	err := webhook.Decode(reader)

	assert.Nil(err)
	assert.Equal(webhook.Ref, "ref/head/test-branch")
}

func TestGithub_Validate_WhenValid(t *testing.T) {
	assert := assert.New(t)
	webhook := baseWebhook()
	ok, err := webhook.Validate("splunk-users-test.yml")

	assert.Nil(err)
	assert.True(ok)
}

func TestGithub_Validate_WrongBranch(t *testing.T) {
	assert := assert.New(t)
	webhook := baseWebhook()
	webhook.Ref = "ref/head/wrong-branch"

	ok, err := webhook.Validate("splunk-users-test.yml")

	assert.NotNil(err)
	assert.False(ok)
}

func TestGithub_Validate_InvalidRepo(t *testing.T) {
	assert := assert.New(t)
	webhook := baseWebhook()
	webhook.Repository.FullName = "heroku/not-access-requests"

	ok, err := webhook.Validate("splunk-users-test.yml")

	if assert.NotNil(err) {
		assert.Equal(err.Error(), "Github webhook from invaliad repository: heroku/not-access-requests")
	}

	assert.False(ok)
}

func TestGithub_Validate_NoModification(t *testing.T) {
	assert := assert.New(t)
	webhook := baseWebhook()
	webhook.HeadCommit.Modified = []string{}

	ok, err := webhook.Validate("splunk-users-test.yml")

	assert.Nil(err)
	assert.False(ok)
}

func baseWebhook() *WebhookPayload {
	webhook := new(WebhookPayload)
	webhook.Ref = "ref/head/test-branch"
	webhook.HeadCommit.Modified = []string{"splunk-users-test.yml"}
	webhook.Repository.FullName = "heroku/access-requests"

	return webhook
}
