package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/kthldap"
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
		p := must1(kthldap.Lookup(ctx, kthid))
		if p == nil {
			panic("No such user")
		}
		slog.Info("Adding user", "user", *p)
		assert(db.CreateUser(ctx, database.CreateUserParams{
			Kthid:      p.KTHID,
			UgKthid:    p.UGKTHID,
			Email:      p.KTHID + "@kth.se",
			FirstName:  p.FirstName,
			FamilyName: p.FamilyName,
			YearTag:    "D" + time.Now().Format("06"),
			MemberTo:   pgtype.Date{Time: time.Now().AddDate(1, 0, 0), Valid: true},
		}))
	case "goose":
		_, db := must2(database.Connect(ctx))
		gooseCMD := shift()
		must0(goose.RunContext(context.Background(), gooseCMD, db(), "database/migrations", args...))
	case "gen-oidc-provider-key":
		key := must1(rsa.GenerateKey(rand.Reader, 4096))
		if len(key.Primes) != 2 {
			panic("I wanted to primes but got " + strconv.Itoa(len(key.Primes)))
		}
		if key.E != 65537 {
			panic("Expected public expontent e to be the standard 65537 but got " + strconv.Itoa(key.E))
		}
		fmt.Printf(
			"OIDC_PROVIDER_KEY=%s,%s\n",
			key.Primes[0].Text(62),
			key.Primes[1].Text(62),
		)
	default:
		panic("No such subcommand")
	}
}

func must0(err error) {
	if err != nil {
		panic(err)
	}
}

func must1[T any](t T, err error) T {
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
