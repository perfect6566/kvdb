module kvdb.com

go 1.13

require (
	github.com/gin-gonic/gin v1.6.3 // indirect
	kvdb.com/Configs v1.0.0
	kvdb.com/kvdbimp v1.0.0
)

replace kvdb.com/kvdbimp => ./kvdbimp

replace kvdb.com/Configs => ./Configs
