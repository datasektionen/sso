package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/kthldap"
	"github.com/jackc/pgx/v5/pgtype"
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
		db := must(database.Connect(ctx))
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
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
