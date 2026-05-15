.PHONY: lib-linux lib-darwin lib-all clean help

lib-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -o libstunning.so ./clib/

lib-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -buildmode=c-shared -o libstunning.dylib ./clib/
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -buildmode=c-shared -o libstunning-arm64.dylib ./clib/

lib-all: lib-linux lib-darwin

clean:
	rm -f libstunning.so libstunning.dylib libstunning-arm64.dylib libstunning.h

help:
	@echo "Build C library: make lib-linux|lib-darwin|lib-all"
	@echo "Clean: make clean"
