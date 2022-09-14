BINARY=ipamtool

.PHONY: build

build:
	@go build -o ${BINARY}