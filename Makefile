.PHONY: help run test retest coverage build build_pkgs clean tools deps update_deps dist

all:
	@echo "********************"
	@echo "** QMD build tool **"
	@echo "********************"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  run         - run in dev mode"
	@echo "  test        - run go tests"
	@echo "  retest      - run go tests, guard style"
	@echo "  build       - build binaries into bin/ directory"
	@echo "  clean       - clean up bin/ directory"
	@echo ""
	@echo "  dist        - clean build with deps and tools"
	@echo "  tools       - go get's a bunch of tools for dev"
	@echo "  deps        - pull and setup dependencies"
	@echo "  update_deps - update deps lock file"

run:
	@test -f ./etc/qmd.conf || { echo "./etc/qmd.conf file missing"; exit 1; }
	@cd ./cmd/qmd && CONFIG=../../etc/qmd.conf fresh -w=../..

test:
	@go test ./...

retest: test
	reflex -r "^*\.go$$" -- make test

coverage:
	@go test -cover -v ./...

build:
	@mkdir -p ./bin
	go build -o ./bin/qmd github.com/pressly/qmd/cmd/qmd

clean:
	@rm -rf ./bin

tools:
	go get github.com/robfig/glock
	go get github.com/cespare/reflex
	go get github.com/pkieltyka/fresh

deps:
	glock sync -n github.com/pressly/qmd < Glockfile

update_deps:
	glock save -n github.com/pressly/qmd > Glockfile

dist: clean tools deps build
