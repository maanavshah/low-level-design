package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Shards []ShardedPool

var ConnectionPool Shards

type ShardedPool struct {
	Master  *sql.DB
	Replica []*sql.DB
}

func createConnection(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	return db
}

func (s *Shards) GetShard(key string) ShardedPool {
	shardId := int(key[0]) % len(ConnectionPool)
	fmt.Println(fmt.Sprintf("using shard: %d", shardId))
	return ConnectionPool[shardId]
}

func (s *Shards) Query(key, query string) (*sql.Rows, error) {
	shard := ConnectionPool.GetShard(key)
	replica := shard.GetReplica()
	return replica.Query(query)
}

func (s *Shards) Exec(key, query string) (sql.Result, error) {
	shard := ConnectionPool.GetShard(key)
	return shard.Master.Exec(query)
}

func (sp *ShardedPool) GetReplica() *sql.DB {
	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(len(sp.Replica)) % len(sp.Replica)
	fmt.Println(fmt.Sprintf("using replica: %d", randomInt))
	return sp.Replica[randomInt]
}

func NewShardedPool(masterDsn string, replicaDsn ...string) ShardedPool {
	var shardedPool ShardedPool
	shardedPool.Master = createConnection(masterDsn)
	for _, dsn := range replicaDsn {
		shardedPool.Replica = append(shardedPool.Replica, createConnection(dsn))
	}
	return shardedPool
}

func main() {
	shard1 := NewShardedPool("root:@/testdb", "root:@/testdb", "root:@/testdb", "root:@/testdb", "root:@/testdb")
	shard2 := NewShardedPool("root:@/testdb", "root:@/testdb", "root:@/testdb", "root:@/testdb", "root:@/testdb")
	ConnectionPool = []ShardedPool{shard1, shard2}

	// Example query
	rows, err := ConnectionPool.Query("m,xampleKey1", "SELECT * FROM table_name")
	if err != nil {
		fmt.Println(fmt.Sprintf("Error in connection pool query: %v", err))
	} else {
		// Process rows
		defer rows.Close()
		for rows.Next() {
			// Handle rows
		}
	}

	// Example execution
	_, err = ConnectionPool.Exec("exampleKey1", "UPDATE table_name SET id = 1;")
	if err != nil {
		fmt.Println(fmt.Sprintf("Error in connection pool exec: %v", err))
	}
}
