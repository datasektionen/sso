package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
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
		p := must1(kthldap.Lookup(ctx, kthid))
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
		must0(goose.RunContext(context.Background(), gooseCMD, db(), "pkg/database/migrations", args...))
	case "gen-oidc-provider-key":
		key := must1(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))
		var j struct {
			X *big.Int `json:"X"`
			Y *big.Int `json:"Y"`
			D *big.Int `json:"D"`
		}
		j.X = key.X
		j.Y = key.Y
		j.D = key.D
		fmt.Println("OIDC_PROVIDER_KEY="+string(must1(json.Marshal(j))))
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
