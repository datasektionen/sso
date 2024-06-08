package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/kthldap"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pressly/goose/v3"
)

func main() {
	args := os.Args[1:]
	shift := func() string {
		if len(args) == 0 {
			panic("Expected another argument")
		}
		s := args[0]
		args = args[1:]
		return s
	}

	ctx := context.Background()
	switch shift() {
	case "add-user":
		db, _ := must2(database.Connect(ctx))
		kthid := shift()
		p, err := kthldap.Lookup(ctx, kthid)
		if err != nil {
			panic(err)
		}
		if p == nil {
			panic("No such user")
		}
		slog.Info("Adding user", "user", *p)
		assert(db.CreateUser(ctx, database.CreateUserParams{
			Kthid:     p.KTHID,
			UgKthid:   p.UGKTHID,
			Email:     p.KTHID + "@kth.se",
			FirstName: p.FirstName,
			Surname:   p.Surname,
			YearTag:   "D" + time.Now().Format("06"),
			MemberTo:  pgtype.Date{Time: time.Now().AddDate(1, 0, 0), Valid: true},
		}))
	case "goose":
		_, db := must2(database.Connect(ctx))
		gooseCMD := shift()
		err := goose.RunContext(context.Background(), gooseCMD, db(), "pkg/database/migrations", args...)
		if err != nil {
			panic(err)
		}
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func must2[T any, Y any](t T, y Y, err error) (T, Y) {
	if err != nil {
		panic(err)
	}
	return t, y
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
