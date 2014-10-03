.PHONY: all

all:
	@echo "********************"
	@echo "** QMD build tool **"
	@echo "********************"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  run         - run the program in dev mode"
	@echo "  test        - standard go test"
	@echo "  convey      - TDD runner"
	@echo "  build       - build the dist binary"
	@echo "  clean       - clean the dist build"
	@echo ""
	@echo "  tools       - go get's a bunch of tools for dev"
	@echo "  deps        - pull and setup dependencies"
	@echo "  update_deps - update deps lock file"

run:
	@(CONFIG=$$PWD/qmd.conf export `goop env` && \
		cd ./cmd/qmd-server && \
		goop exec fresh -w=../..)

test:
	@goop go test

convey:
	@goop exec goconvey

build:
	@mkdir -p ./bin
	@rm -f ./bin/*
	goop go build -o ./bin/qmd-server github.com/pressly/qmd/cmd/qmd-server

clean:
	@rm -rf ./bin

tools:
	go get github.com/pkieltyka/goop
	go get github.com/pkieltyka/fresh
	go get github.com/smartystreets/goconvey/...

deps:
	@rm -rf ./.vendor
	goop install

update_deps:
	goop update
