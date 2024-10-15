APP := CompareFK
PACKAGE := main

REVISION := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
DATE := $(shell date +%F)

build_win:
	GOOS=windows GOARCH=amd64 go build  -ldflags "-X $(PACKAGE).buildCommit=$(REVISION) -X $(PACKAGE).buildVersion=$(BRANCH) -X $(PACKAGE).buildDate=$(DATE)" -o bin/$(APP).exe  cmd/compare/main.go

build:
	go build  -ldflags "-X $(PACKAGE).buildCommit=$(REVISION) -X $(PACKAGE).buildVersion=$(BRANCH) -X $(PACKAGE).buildDate=$(DATE)" -o bin/$(APP)  cmd/compare/main.go