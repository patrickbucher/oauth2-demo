# oauth2-demo

These are some toy applications to demonstrate OAuth 2. **Do not reuse the
security mechanisms demonstrated here in productive applications!** The code is
not so much about demonstrating security best practices, but to elucidate the
flow between the components in OAuth 2.

The oauth2-demo consists of three services to serve gossip.

TODO: describe endpoints and flow in more detail

## Components

### Client (Port 1234)

The `client` fetches the resource owner's gossip from the `resource`.

#### `/`

Shows the input form to enter a username (scope) for the gossip to retrieve.

#### `/gossip`

Endpoint to submit the form to.

- method: `POST`
- form fields: `username` (possible values: alice, bob, mallory)

#### `/callback/[scope]`

Endpoint for the `authserver` to get back to the client after authorisation.

- method: `GET`
- query parameters:
    - `auth_host`: hostname or IP address of the authorization server (`localhost`)
    - `auth_port`: port number of the authorization server (`8443`)
    - `auth_code`: authorization code, a one-time password (random, base64 encoded string)
    - `state`: request identifier initially generated from client (random, base64 encoded string)

### Resource (Port 8000)

The `resource` holds the resource owner's gossip and serves it if a valid access token is used.

#### `/gossip/[scope]`

Endpoint to serve the gossip of a certain user (username = scope).

- method: `POST`
- request headers:
    - `Authorization`: authorization header with a value of the form `Bearer [access_token]`
- query parameters:
    - `host`: hostname of IP address of the client (`localhost`)
    - `port`: port number of the client (`1234`)
    - `client_id`: the client's ID
    - `state`: request identifier initially generated from client (random, base64 encoded string)
- response headers (if redirected due to missing/invalid `access_token`):
    - `WWW-Authenticate: bearer`
    - `Location: [redirect_url]`
        - with a redirect URL like `http://localhost:8443/authorization?callback_url=[callback_url]`
        - with a (URL encoded) callback URL like `http://localhost:1234/callback/alice?state=20JDxzVi2MlfCa6K8323tQ&client_id=gossip_client`
- response body (if valid `access_token`): JSON-encoded gossip

Example response body:

```json
[
  "Oreos are made out of sand.",
  "Bob stinks."
]
```

### Authserver (Port 8443)

The `authserver` handles user and client authentication, the user's client authorisations and manages the access tokens.

#### `/authorization`

Endpoint that lets a user authenticate himself and authorize a client.

- `GET`: show authentication form
    - query parameter: `callback_url` (as shown above)
- `POST`: submit authentication form
    - query parameters:
        - `username` and `password` (entered manually)
        - `callback_url` (hidden form field)
    - response headers:
        - `Location: [redirect_url`], the `callback_url` above with additional parameters (`auth_host`, `auth_port`, `auth_code`; as documented in the client section)

#### `/token`

Endpoint that provides an access token in exchange of a valid authorization code and valid client credentials.

- method: `POST`
- request headers:
    - `Authorization: Basic [client_id:client_secret]`, with base64 encoded `client_id` and `client_secret`
- form fields:
    - `grant_type=authorization_code` (constant value, the only supported grant type)
    - `authorization_code=[authorization_code]` (as described above)
- response headers (if client credentials and authorization code are valid):
    - `Content-Type: application/json`
- response body (valid client credentials and authorization code): JSON-encoded access token

Example response body:

```json
{
    "access_token": "c2hpaGFlTmdhaXM3SWV3aWVQdTJvaHNlZVZlR2Vld28K",
    "token_type": "Bearer"
}
```

#### `/accesscheck`

Endpoint that checks if a submitted access token is (still) valid.

- method: `POST`
- form fields:
    - `access_token`: the access token in question
    - `scope`: the scope the access token is supposedly valid for
- response:
    - `200 OK` if the token is valid
    - `403 Forbidden` if the token is invalid

## Run the Demo Applications

Requirements: Go version >= 1.11

On Linux/macOS: using the script `run.sh`

On Windows: by performing the steps in `run.sh` manually
