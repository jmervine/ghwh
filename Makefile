GOVENDOR=$(shell echo "$(GOBIN)/govendor")
GODOTENV=$(shell echo "$(GOBIN)/godotenv")

test: $(GODOTENV) $(GOVENDOR)
	godotenv -f .env.test govendor test +local

ci: $(GOVENDOR)
	govendor test +local

$(GODOTENV):
	go get -v github.com/joho/godotenv/cmd/godotenv

$(GOVENDOR):
	go get -v github.com/kardianos/govendor
