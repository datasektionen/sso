## Routes

```sh
grep -r 'mux.Handle''(' --no-filename . | sed 's/^\s\+//'
```

### API

An OpenID Connect provider is available at `/op`, (so the production configuration endpoint is at
`https://sso.datasektionen.se/op/.well-known/openid-configuration`). In addition to some standard
user info stuff, SSO supports some custom scopes:

- `pls_$SYSTEM`: will make the returned claims have a key of the same name (including `pls_`) which
  is an array of strings which are the permissions in [pls](https://github.com/datasektionen/pls)
  for the given system the user has. This is now considered and the `permissions` scope should be
  fetched instead.
- `permissions`: will make the returned claims include a key of the same name whose value is the
  user's permissions in [Hive](https://github.com/datasektionen/hive). The system must be specified
  for each OIDC client in SSO's admin panel. The value follows the same format as the body in the
  [`/user/{username}/permissions`](https://hive.datasektionen.se/api/v1/docs#get-/user/-username-/permissions)
  endpoint in Hive's API.

---

`/legacyapi`: follows the same API as [login2](https://github.com/datasektionen/login).

#### Internal API

These are only reachable from other systems, on the domain name `sso.nomad.dsekt.internal`. This
means that you cannot call these from code that runs in the browser.

`GET /api/users`: Retrieves user information (email, first name, family name, year tag) by their
usernames. The url query parameter `u`, which can be repeated, determines which users to retrieve.
The url query parameter `format` determines how the users are returned:
- `single`: exactly one must be requested. If it is not found, the status code will be 404,
  otherwise the response body will contain a json object with the keys `email`, `firstName`,
  `familyName`, `yearTag`. If any key (mostly this will be `yearTag`) is missing/empty, it will not
  be present.
- `array`: The response body will contain a json array with objects as explained in `single`. Their
  order will always be the same as the usernames in the request (which must not be repeated).
  Requested non-existing users are represented by empty json objects.
- `map`: The response body will contain a json object with keys being usernames and values being
  objects as explained in `single`. Requested non-existing users will not be present in the response.

## Development

### Install dependencies

Download a go compiler, at least version 1.24

If you will modify any `.sql` files, download [sqlc](https://sqlc.dev/). It's
probably best to get the latest version, using:
```sh
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

But to see which version was last used, look at the top of any generated file, e.g. `database/models.go`.

If you will modify `style.css` or any `.templ` files you need to download [templ](https://templ.guide/) and [tailwind](https://tailwindcss.com/).

Download the correct version of templ using:
```sh
go install github.com/a-h/templ/cmd/templ@$(grep -oPm1 'github.com/a-h/templ \K[^ ]*' go.sum)
```

(It should be the same version as is imported by the then generated go code)

Download tailwind, using npm, your favourite or second-favourite package manager, or:
```sh
TAILWIND_VERSION=$(head -n2 pkg/static/public/style.dist.css | grep -oP 'v\K\S+') # or use latest, it probably won't hurt.
curl -Lo tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/v$TAILWIND_VERSION/tailwindcss-linux-x64
chmod +x tailwindcss
```
and then move it to a directory in your `$PATH` or set `$TAILWIND_PATH` to it's location.

### Database

Start a postgresql database, using e.g.:
```sh
docker run -d --name sso-db -p 5432:5432 -e POSTGRES_PASSWORD=sso -e POSTGRES_DB=sso -e POSTGRES_USER=sso docker.io/postgres:16-alpine3
```
...or add a user and database to an existing instance:
```sql
CREATE USER sso WITH PASSWORD 'sso';
CREATE DATABASE sso WITH OWNER sso;
```

Then you need to run the migrations on the database, which can be done by running
`go run ./cmd/manage goose up`, or just starting the application.

### Environment variables

Copy `example.env` to `.env` and change anything if needed.

The best way (objectively, of course) to load them is by installing
[direnv](https://direnv.net/).

If you don't want to do that, you'll have to figure it out on your own. The
application will not care about the `.env` file itself.

### Creating a first user

```sh
go run ./cmd/manage add-user turetek
```

### Running with automatic recompiling & rerunning

```sh
go run ./cmd/dev
```

This also updates generated files and they're tracked by the git repository so you better.

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
vault write identity/oidc/client/sso redirect_uris="http://localhost:7000/oidc/kth/callback" assignments=allow_all

echo -n "
KTH_ISSUER_URL=$(vault read -field=issuer identity/oidc/provider/default)
KTH_CLIENT_ID=$(vault read -field=client_id identity/oidc/client/sso)
KTH_CLIENT_SECRET=$(vault read -field=client_secret identity/oidc/client/sso)
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

## Database schema

The schema is defined by the migrations in `./database/migrations/`. A new
one can be created using `go run ./cmd/manage goose create some_fancy_name sql`.
