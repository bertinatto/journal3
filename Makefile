all:
	@go build -o ./.output/journal3 ./cmd/journal3

db:
	@sqlite3 blog.db < schema.sql

image:
	@podman build . -t quay.io/bertinatto/journal3

push:
	@podman push quay.io/bertinatto/journal3
