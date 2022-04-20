NAME := new-release-version
ORG := trendmicro
VERSION := 1.0

all: build test

.PHONY: build
build:
	go build -v .

.PHONY: test
test:
	go test -v .

.PHONY: clean
clean:
	-$(RM) $(NAME)
