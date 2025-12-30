package database

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/nanoteck137/authlab/types"
)

func addToRecord[T any](record goqu.Record, name string, change types.Change[T]) {
	if change.Changed {
		record[name] = change.Value
	}
}
