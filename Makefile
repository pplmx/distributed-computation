.PHONY: help wire gen image run build test
.DEFAULT_GOAL := help

APP_NAME=distributed-computation
APP_PATH=.

# run
run:
	@go run $(APP_PATH)/main.go

# build
build:
	@mkdir -p bin && go mod tidy && go build -trimpath -ldflags="-w -s" -o bin/ $(APP_PATH)

# test
test:
	@go test -v ./...

# fmt
fmt:
	@goimports -l -w .

# update my own pb
up:
	@go get -u github.com/pplmx/pb/dist

# Show help
help:
	@echo ""
	@echo "Usage:"
	@echo "    make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z\-_0-9]+:/ \
	{ \
		helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} { lastLine = $$0 }' $(MAKEFILE_LIST)
