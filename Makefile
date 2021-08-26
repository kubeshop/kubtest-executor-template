.PHONY: test cover 

BIN_DIR ?= $(HOME)/bin
GITHUB_TOKEN ?= "SET_ME"
USER ?= $(USER)
NAMESPACE ?= "kt1"
DATE ?= $(shell date -u --iso-8601=seconds)
COMMIT ?= $(shell git log -1 --pretty=format:"%h")

run-executor: 
	CYPRESSEXECUTOR_PORT=8083 go run cmd/cypress-executor/main.go

run-mongo-dev: 
	docker run -p 27017:27017 mongo


build: 
	go build -o $(BIN_DIR)/cypress-executor cmd/cypress-executor/main.go




# build done by vendoring to bypass private go repo problems
docker-build-executor: 
	go mod vendor
	docker build --build-arg TOKEN=$(GITHUB_TOKEN) -t cypress-executor -f build/cypress-executor/Dockerfile .

install-swagger-codegen-mac: 
	brew install swagger-codegen

test: 
	go test ./... -cover

test-e2e:
	go test --tags=e2e -v ./test/e2e

test-e2e-namespace:
	NAMESPACE=$(NAMESPACE) go test --tags=e2e -v  ./test/e2e 

cover: 
	@go test -failfast -count=1 -v -tags test  -coverprofile=./testCoverage.txt ./... && go tool cover -html=./testCoverage.txt -o testCoverage.html && rm ./testCoverage.txt 
	open testCoverage.html


version-bump: version-bump-patch

version-bump-patch:
	go run scripts/bump.go -kind patch

version-bump-minor:
	go run scripts/bump.go -kind minor

version-bump-major:
	go run scripts/bump.go -kind major