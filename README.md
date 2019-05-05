# oauth2-demo

These are some toy applications to demonstrate OAuth 2. **Do not reuse the
security mechanisms demonstrated here in productive applications!** The code is
not so much about demonstrating security best practices, but to elucidate the
flow between the components in OAuth 2.

The oauth2-demo consists of three services to serve gossip.

TODO: describe endpoints and flow in more detail

## Components

### `client` (Port 1234)

The `client` fetches the resource owner's gossip from the `resource`.

#### `/`

Shows the input form to enter a username (scope) for the gossip to retrieve.

#### `/gossip`

Endpoint to submit the form to.

#### `/callback/[scope]`

Endpoint for the `authserver` to get back to the client after authorisation.

### `resource` (Port 8000)

The `resource` holds the resource owner's gossip and serves it if a valid access token is used.

#### `/gossip/[scope]`

Endpoint to serve the gossip of a certain user (username = scope).

### `authserver` (Port 8443)

The `authserver` handles user and client authentication, the user's client authorisations and manages the access tokens.

#### `/authorization`

Endpoint that lets a user authenticate himself and authorize a client.

- `GET`: show authentication form
- `POST`: submit authentication form

#### `/token`

Endpoint that provides an access token in exchange of a valid authorization code and valid client credentials.

#### `/accesscheck`

Endpoint that checks if a submitted access token is (still) valid.

## Run the Demo Applications

TODO: using `docker-compose` (all platforms)

TODO: using the `run.sh` script (Linux, and maybe macOS)
