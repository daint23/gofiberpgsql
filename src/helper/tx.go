package helper

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func CommitOrRollback(tx pgx.Tx) {
	err := recover()
	if err != nil {
		errRoll := tx.Rollback(context.Background())
		if errRoll != nil {
			fmt.Println("errRol")
			panic(errRoll)
		}
		fmt.Println("errAja")
		panic(err)
	} else {
		errComm := tx.Commit(context.Background())
		if errComm != nil {
			fmt.Println("errComm")
			panic(errComm)
		}
	}
}
