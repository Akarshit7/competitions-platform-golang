.PHONY: build run


build-api:
	go build -o bin/api  cmd/api/main.go

run-api: build-api
	./bin/api

clean:
	rm -f bin/api

