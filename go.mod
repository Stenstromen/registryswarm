module github.com/stenstromen/registryswarm

go 1.21

replace github.com/stenstromen/registryswarm => /

require (
	github.com/mattn/go-sqlite3 v1.14.18
	gopkg.in/yaml.v2 v2.4.0
)

require github.com/gorilla/mux v1.8.1 // indirect
