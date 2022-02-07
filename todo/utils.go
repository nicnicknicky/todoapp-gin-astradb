package todo

import "github.com/gocql/gocql"

func mustParseUUID(s string) gocql.UUID {
	uuid, err := gocql.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return uuid
}
