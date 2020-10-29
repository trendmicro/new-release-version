NAME := new-release-version
ORG := trendmicro
VERSION := 2.1

all: build test

.PHONY: test
test:
	go test -v .

.PHONY: build
build:
	go build -v .

.PHONY: clean
clean:
	-$(RM) $(NAME)
