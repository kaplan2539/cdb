VERSION="0.0"
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE=$(shell date --iso-8601)
VERSION_FILE=libcdb/version.go

## This is an arbitrary comment to arbitrarily change the commit hash

GOPATH?=$(shell go env GOPATH)

CDB_CLI_SOURCES=$(shell ls cdb-cli/*.go)
CDB_DAEMON_SOURCES=$(shell ls cdb-daemon/*.go)
LIBCDB_SOURCES=$(shell ls libcdb/*.go)

DEPENDS=\
	github.com/nextthingco/logrus-gadget-formatter

## Bottom two libs here^^ are essentially one-off code chunks which aren't
## likely to be updated. Neither has tags, and thus, gopkg.in links aren't
## being used.

all: cdb cdbd

cdb: libcdb $(CDB_CLI_SOURCES) $(VERSION_FILE) $(LIBCDB_SOURCES)
	@echo "Building cdb-cli"
	@mkdir -p build/linux_arm
	@mkdir -p build/linux_arm64
	@mkdir -p build/linux_amd64
	@GOOS=linux GOARCH=arm   go build -o build/linux_arm/cdb -ldflags="-s -w" -v ./cdb-cli
	@GOOS=linux GOARCH=arm64 go build -o build/linux_arm64/cdb -ldflags="-s -w" -v ./cdb-cli
	@GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/cdb -ldflags="-s -w" -v ./cdb-cli

$(VERSION_FILE):
	@echo "package libcdb" > $(VERSION_FILE)
	@echo "const (" >> $(VERSION_FILE)
	@echo "	Version = \"${VERSION}\"" >> $(VERSION_FILE)
	@echo "	GitCommit = \"${GIT_COMMIT}\"" >> $(VERSION_FILE)
	@echo "	GitBranch = \"${GIT_BRANCH}\"" >> $(VERSION_FILE)
	@echo "	BuildDate = \"${BUILD_DATE}\"" >> $(VERSION_FILE)
	@echo ")" >> $(VERSION_FILE)


cdbd: libcdb $(CDB_DAEMON_SOURCES) $(VERSION_FILE) $(LIBCDB_SOURCES)
	@echo "Building cdb-daaemon"
	@mkdir -p build/linux_arm
	@mkdir -p build/linux_arm64
	@mkdir -p build/linux_amd64
	@GOOS=linux GOARCH=arm   go build -o build/linux_arm/cdbd -ldflags="-s -w" ./cdb-daemon
	@GOOS=linux GOARCH=arm64 go build -o build/linux_arm64/cdbd -ldflags="-s -w" ./cdb-daemon
	@GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/cdbd -ldflags="-s -w" ./cdb-daemon

libcdb: $(VERSION_FILE)
	@echo "Building libcdb"
	@rm -rf ${GOPATH}/src/github.com/nextthingco/libcdb
	@cp -r libcdb ${GOPATH}/src/github.com/nextthingco/
	@go install -ldflags="-X libcdb.Version=$(VERSION) -X libcdb.GitCommit=$(GIT_COMMIT)" -v github.com/nextthingco/libcdb

tidy:
	@echo "Tidying up sources"
	@go fmt ./cdb-cli
	@go fmt ./cdb-daemon
	@go fmt ./libcdb

clean:
	@echo "Cleaning"
	@rm -rf build/ $(VERSION_FILE)

test: $(CDB_CLI_SOURCES) $(CDB_CLI_SOURCES)
	@echo "Testing CDB"
	@go test -ldflags="-s -w" -v ./cdb-cli
	@go test -ldflags="-s -w" -v ./libcdb

get:
	@echo "Downloading external dependencies"
	@go get ${DEPENDS}
