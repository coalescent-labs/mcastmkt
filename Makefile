BINARY=mcastmkt
PLATFORMS=linux windows
ARCHITECTURES=amd64
PACKAGE=github.com/coalescent-labs/mcastmkt
VERSION?=0.0.0

# setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-w -s -X '${PACKAGE}/pkg/version.version=${VERSION}' -X '${PACKAGE}/pkg/version.commit=${GIT_COMMIT}' -X '${PACKAGE}/pkg/version.branch=${GIT_BRANCH}'"

default: mod vet fmt build

all: clean mod test vet fmt build_all

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

mod:
	go mod tidy
	go mod vendor

build:
	CGO_ENABLED=0 go build ${LDFLAGS} -o bin/$(BINARY)

build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build ${LDFLAGS} -v -o bin/$(BINARY)-$(GOOS)-$(GOARCH))))

run:
	chmod +x bin/mcastmkt
	./bin/mcastmkt

install:
	go install -v ./...

clean:
	rm -f bin/*

.PHONY: default all test vet fmt mod build build_all run install clean