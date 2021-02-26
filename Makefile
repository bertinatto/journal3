all:
	@go build ./cmd/journal3

db:
	@sqlite3 blog.db < schema.sql
