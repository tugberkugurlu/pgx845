package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/jackc/pgx/v4"
)

const setupSQL = `drop table if exists j;

create table j(
	data jsonb NOT NULL
);

insert into j(data) values ('{ "phones":[ {"type": "mobile", "phone": "001001"} , {"type": "fix", "phone": "002002"} ] }');
insert into j(data) values ('{ "phones":[ {"type": "mobile-2", "phone": "001001"} , {"type": "fix-2", "phone": "002002"} ] }');
insert into j(data) values ('{ "phones":[ {"type": "mobile-3", "phone": "001001"} , {"type": "fix-3", "phone": "002002"} ] }');
`

type Foo struct {
	Phones []struct {
		Type  string `json:"type"`
		Phone string `json:"phone"`
	} `json:"phones"`
}

func main() {
	var memStats runtime.MemStats

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln("Unable to connect to database:", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), setupSQL)
	if err != nil {
		log.Fatalln("Unable to setup database:", err)
	}

	for i := 0; i < 1000000000; i++ {
		func() {
			rows, err := conn.Query(context.Background(), "select data from j;")
			if err != nil {
				log.Fatalln("conn.Query unexpectedly failed:", err)
			}
			defer rows.Close()

			var result []*Foo
			for rows.Next() {
				var data Foo
				scanErr := rows.Scan(&data)
				if scanErr != nil {
					log.Fatalln("rows.Scan unexpectedly failed:", scanErr)
				}
				result = append(result, &data)
			}
			if err := rows.Err(); err != nil {
				log.Fatalln("rows.Err() is not nil:", err)
			}
			if resLen := len(result); resLen != 3 {
				log.Fatalln("len(result) is expected to be 3, but got", resLen)
			}

			if i%100000 == 0 {
				runtime.GC()
				runtime.ReadMemStats(&memStats)
				fmt.Printf("i=%d\tHeapAlloc=%d\tHeapObjects=%d\n", i, memStats.HeapAlloc, memStats.HeapObjects)
			}
		}()
	}
}
