// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: oidcprovider.sql

package database

import (
	"context"
)

const createClient = `-- name: CreateClient :one
insert into oidc_clients (id, secret_hash, redirect_uris)
values ($1, $2, '{}')
returning secret_hash, redirect_uris, id
`

type CreateClientParams struct {
	ID         string
	SecretHash []byte
}

func (q *Queries) CreateClient(ctx context.Context, arg CreateClientParams) (OidcClient, error) {
	row := q.db.QueryRow(ctx, createClient, arg.ID, arg.SecretHash)
	var i OidcClient
	err := row.Scan(&i.SecretHash, &i.RedirectUris, &i.ID)
	return i, err
}

const deleteClient = `-- name: DeleteClient :exec
delete from oidc_clients
where id = $1
`

func (q *Queries) DeleteClient(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteClient, id)
	return err
}

const getClient = `-- name: GetClient :one
select secret_hash, redirect_uris, id
from oidc_clients
where id = $1
`

func (q *Queries) GetClient(ctx context.Context, id string) (OidcClient, error) {
	row := q.db.QueryRow(ctx, getClient, id)
	var i OidcClient
	err := row.Scan(&i.SecretHash, &i.RedirectUris, &i.ID)
	return i, err
}

const listClients = `-- name: ListClients :many
select secret_hash, redirect_uris, id
from oidc_clients
`

func (q *Queries) ListClients(ctx context.Context) ([]OidcClient, error) {
	rows, err := q.db.Query(ctx, listClients)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OidcClient
	for rows.Next() {
		var i OidcClient
		if err := rows.Scan(&i.SecretHash, &i.RedirectUris, &i.ID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateClient = `-- name: UpdateClient :one
update oidc_clients
set redirect_uris = $2
where id = $1
returning secret_hash, redirect_uris, id
`

type UpdateClientParams struct {
	ID           string
	RedirectUris []string
}

func (q *Queries) UpdateClient(ctx context.Context, arg UpdateClientParams) (OidcClient, error) {
	row := q.db.QueryRow(ctx, updateClient, arg.ID, arg.RedirectUris)
	var i OidcClient
	err := row.Scan(&i.SecretHash, &i.RedirectUris, &i.ID)
	return i, err
}
