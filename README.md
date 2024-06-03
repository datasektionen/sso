## Repository structure

This repository contains two main parts: a Go web server and some solid
components written in typescript. The web server in Go does most of the job, it
has a few API:s returning JSON but mostly server-rendered routes that the
browser interacts with directly (so it's not just a backend). Some parts of the
page however require a bit more interactivity; these are made with
[solid](https://www.solidjs.com/) (which is like react but less cringe™). This
is sometimes called _interactive islands_.

### Solid

These are build with [rollup](https://rollupjs.org/) for bundling,
[babel](https://babeljs.io/) for building towers or whatever (the solid
compiler is a babel plugin), typescript (to not loose sanity) and solid
(through babel-preset-solid). They are configured with `rollup.config.mjs` and
`tsconfig.json`.

For each file ending with `.island.tsx` in the `islands/` directory a
corresponding file ending in `.island.js` will be placed in `dist/` which the
Go server will serve at `/dist`. These can be loaded using a script tag with
`type="module"`.

Also note that `pnpm` is used, not `npm`. You can install `pnpm` by enabling
[corepack](https://nodejs.org/api/corepack.html).

## Routes

```sh
grep -r 'http.Handle(' --no-filename services/ | sed 's/\s\+//'
```

## Development

### Install dependencies

Download a go compiler, at least version 1.22

Download the correct version of [templ](https://templ.guide/) using:
```sh
go install github.com/a-h/templ/cmd/templ@$(grep -oPm1 'github.com/a-h/templ \K[^ ]*' go.sum)
```

Download [sqlc](https://sqlc.dev/). It's probably best to get the latest version, using:
```sh
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

But to see which version was last used, look at the top of any generated file, e.g. `pkg/database/models.go`.

### Database

Start a postgresql database, using e.g.:
```sh
docker run -d --name logout-db -p 5432:5432 -e POSTGRES_PASSWORD=logout -e POSTGRES_DB=logout -e POSTGRES_USER=logout postgres:16-alpine3.19
```
...or add a user and database to an existing instance:
```sql
CREATE USER logout WITH PASSWORD 'logout';
CREATE DATABASE logout WITH OWNER logout;
```

### Environment variables

Copy `example.env` to `.env` and change anything if needed.

The best way (objectively, of course) to load them is by installing
[direnv](https://direnv.net/).

### Running with automatic recompiling & rerunning

```sh
find -regex '.*\.\(templ\|sql\)' | entr go generate ./...
```

```sh
find -name '*.go' | entr -r go run ./cmd/web
```

```sh
pnpm build --watch
```

### Mocking an OIDC provider

Only needed if you want to test the "Login with KTH" button.

Start a vault dev server (tested with version v1.16.2):

```sh
vault server -dev
```

In another terminal, set it up with the following:

```sh
export VAULT_ADDR=http://127.0.0.1:8200

vault auth enable userpass
vault auth tune -listing-visibility=unauth userpass
vault write auth/userpass/users/turetek password=turetek
vault write identity/entity name=turetek
vault write identity/entity-alias \
    name=turetek \
    canonical_id=$(vault read -field=id identity/entity/name/turetek) \
    mount_accessor=$(vault auth list -detailed -format json | jq -r '.["userpass/"].accessor')

vault write identity/oidc/scope/profile template=$(echo '{"username":{{identity.entity.name}}}' | base64 -)
vault write identity/oidc/provider/default scopes_supported="profile"
vault write identity/oidc/client/logout redirect_uris="http://localhost:7000/oidc/kth/callback" assignments=allow_all

echo -n "
KTH_ISSUER_URL=$(vault read -field=issuer identity/oidc/provider/default)
KTH_CLIENT_ID=$(vault read -field=client_id identity/oidc/client/logout)
KTH_CLIENT_SECRET=$(vault read -field=client_secret identity/oidc/client/logout)
KTH_RP_ORIGIN=http://localhost:7000
"
```

Then set the env variables printed at the end. Then you can log in using
`turetek` as username and password when you've registered `turetek` in the
system. Note that it doesn't mock all the claims we get from kth.

## Example claims from KTH:

```js
{
    "affiliation": [
        "member",
        "student"
    ],
    "appid": "bad94f41-8323-4c26-8c59-fb6d6b8384db",
    "apptype": "Confidential",
    "aud": "bad94f41-8323-4c26-8c59-fb6d6b8384db",
    "auth_time": 1717086052,
    "authmethod": "urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport",
    "email": "turetek@kth.se",
    "exp": 1717090261,
    "iat": 1717086661,
    "iss": "https://login.ug.kth.se/adfs",
    "kthid": "u1jwkms6",
    "nbf": 1717086661,
    "scp": "openid",
    "sid": "S-1-1-11-1111111111-1111111111-1111111111-1111111",
    "sub": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
    "unique_name": [
        "Ture Teknokrat",
        "UG\\turetek"
    ],
    "upn": "turetek@ug.kth.se",
    "username": "turetek",
    "ver": "1.0"
}
```

## Cookie [`SameSite`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#samesitesamesite-value) mode

It may seem like most cookies could and should have the `SameSite` attribute
set to `Strict`, but when the user is redirected back from the KTH login page
to `/oidc/kth/callback` that redirects the user further within this system and
then cookies must be sent, but from some local testing it seems they're not
since the user was (indirectly) redirected from KTH. Therefore they're set to
`Lax`.
