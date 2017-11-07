package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/sirupsen/logrus"
)

type github struct {
	Token  string `env:"GITHUB_TOKEN,required"`
	Repo   string `env:"GITHUB_REPO,default=heroku/access-requests"`
	Branch string `env:"GITHUB_BRANCH,default=master"`
}

const AcceptHeader = "application/vnd.github.v3.raw"

var (
	logger *logrus.Entry
	Config github
)

func Init(log *logrus.Entry) {
	logger = log.WithField("package", "github")

	if err := envdecode.Decode(&Config); err != nil {
		logger.Fatal(err)
	}
}

func InvalidBranchError() error {
	return errors.New("Github webhook Ref contains invalid branch, expected " + Config.Branch)
}

// WebhookPayload describes the incoming json payload from the Github webhook.
type WebhookPayload struct {
	Ref        string `json:"ref"`
	HeadCommit struct {
		Modified []string `json:"modified"`
	} `json:"head_commit"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// Validate ensure that the payload's repo, branch and file all match what's
// configured to be valid.
func (w *WebhookPayload) Validate(usersFile string) (bool, error) {
	if !strings.HasSuffix(w.Ref, Config.Branch) {
		return false, InvalidBranchError()
	}

	if w.Repository.FullName != Config.Repo {
		return false, errors.New("Github webhook from invaliad repository: " + w.Repository.FullName)
	}

	for _, mod := range w.HeadCommit.Modified {
		if mod == usersFile {
			return true, nil
		}
	}

	return false, nil
}

// Decode decodes the incoming payload in to a WebhookPayload.
func (w *WebhookPayload) Decode(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, w)
}

// Fetch fetches the file from the github branch and repo as per the
// confiugration.
func Fetch(file string) (body []byte, err error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", Config.Repo, file)

	start := time.Now()

	req, err := http.NewRequest("GET", endpoint, nil)

	if err != nil {
		return
	}

	req.Header.Add("Accept", AcceptHeader)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", Config.Token))

	params := url.Values{}
	params.Add("ref", Config.Branch)

	req.URL.RawQuery = params.Encode()

	logger.WithFields(logrus.Fields{
		"fn":     "Fetch",
		"source": req.URL.String(),
		"at":     "start",
	}).Debug()

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return
	}

	logger.WithFields(logrus.Fields{
		"fn":     "Fetch",
		"source": req.URL.String(),
		"at":     "finish",
		"status": resp.StatusCode,
		"took":   time.Since(start),
	}).Debug()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.New(resp.Status)
		return
	}

	return ioutil.ReadAll(resp.Body)
}
