VERSION := $(shell git describe --tags --abbrev=0)
LDFLAGS = "-X gsa.gov/18f/version.Semver=$(VERSION)"

ifeq ($(shell uname -m),armv7l)
RELEASEPATH = ../release/bin/*
ENVVARS = CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7
else
RELEASEPATH = ../release/bin/linux_arm/*
ENVVARS = CGO_ENABLED=1 GOOS=linux
endif

ifeq ($(OS),Windows_NT)
# choco install mingw (for now, since we need CGO)
GOEXE = go
RELEASEPATH = ../release/bin/windows/*
ENVVARS = CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/c/ProgramData/chocolatey/lib/mingw/tools/install/mingw64/bin/x86_64-w64-mingw32-gcc
else
GOEXE = go
endif

check-install:
	${GOEXE} get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1

check:
	golangci-lint run ./...

all: input-configuration session-counter wifi-hardware-search-cli log-event cache-to-sqlite waterfall

deps:
	${GOEXE} version
	${GOEXE} mod download

input-configuration: deps
	${ENVVARS} GOPATH=$$PWD/../release ${GOEXE} install -ldflags $(LDFLAGS) gsa.gov/18f/cmd/input-initial-configuration

session-counter: deps
	${ENVVARS} GOPATH=$$PWD/../release ${GOEXE} install -ldflags $(LDFLAGS) gsa.gov/18f/cmd/session-counter

wifi-hardware-search-cli: deps
	${ENVVARS} GOPATH=$$PWD/../release ${GOEXE} install -ldflags $(LDFLAGS) gsa.gov/18f/cmd/wifi-hardware-search-cli

log-event: deps
	${ENVVARS} GOPATH=$$PWD/../release ${GOEXE} install -ldflags $(LDFLAGS) gsa.gov/18f/cmd/log-event

cache-to-sqlite: deps
	${ENVVARS} GOPATH=$$PWD/../release ${GOEXE} install -ldflags $(LDFLAGS) gsa.gov/18f/cmd/cache-to-sqlite

waterfall: deps
	${ENVVARS} GOPATH=$$PWD/../release ${GOEXE} install -ldflags $(LDFLAGS) gsa.gov/18f/cmd/waterfall

clean:
	mkdir -p ../release/{bin,pkg}
	chmod -R +w ../release/{bin,pkg}*
	rm -rf ../release/{bin,pkg}

test:
	${GOEXE} test -coverprofile all.out -timeout 45m ./...
