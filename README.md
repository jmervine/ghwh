# ghwh

Simple abstraction of github webhook functionality. Currently, probably not useful
for much aside from the projects I'm using it in.

```golang
import (
    "fmt"
    "net/http"

    "github.com/sirupsen/logrus"
    "github.com/jmervine/ghwh"
)

func main() {
    logger := logrus.New().WithFields("package", "main")

    // Assumes configuration is set via the Environment
    github.Init(logger)

    // Fetch a file
    data, err := github.Fetch("README.md")
    if err != nil {
        logger.Fatal(err)
    }

    fmt.Println(string(data))

    // Listen for a webook
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        webhook := &github.WebhookPayload{}

        defer r.Body.Close()
        err := webhook.Decode(r.Body)

        if err == io.EOF {
            logger.Error(errors.New("Github webhook payload empty"))
            return
        }

        if err != nil {
            logger.Error(err)
            return
        }

        modified, err = webhook.Validate("modified-file.txt")
        if err != nil {
            if err == github.InvalidBranchError() {
                w.WriteHeader(404)
            } else {
                w.WriteHeader(500)
            }
            fmt.Fprintf(w, err.Error())
        }

        fmt.Fprintf(w, "webhook recieved")
    })
}
