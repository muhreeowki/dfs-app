build:
	@go build -o bin/dfs-app

run: build
	@./bin/dfs-app

test:
	@go test ./... -v 
