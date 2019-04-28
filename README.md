# oauth2-demo

These are some toy applications to demonstrate OAuth 2. **Do not reuse the
security mechanisms demonstrated here in productive applications!** The code is
not so much about demonstrating security best practices, but to  elucidate the
flow between the components in OAuth 2.

## Components and Endpoints

TODO: update endpoints according to implementation

- `resource`: the application holding and serving the protected resources (some
  random gossip)
    - port 8000
    - endpoint `/gossip/[username]`: receive the gossip from a given user
- `authserver`: the server that issues and checks access tokens
    - port 8443
    - endpoint GET `/authorization?client_id=[client_id]`: shows an
      authorization form that lets the user authenticate and authorize access
      for the client with the given `client_id`
    - endpoint POST `/authorization` (form params: `client=[client],
      username=[username], password=[password]`): authenticate a user and grant
      access rights to the client, return a `client_secret`
    - endpoint
      `/token?client_id=[client_id]&client_secret=[client_secret]&username=[username]`:
      issue a new `access_token` for an authorized client for the username
      specified
    - endpoint `/accesscheck?acces_token=[access_token]`: checks if a given
      `access_token` is (still) valid
- `client`: the client application that requests access to the protected
  resources
    - port 1234 
    - endpoint `/index.html`: web frontend that asks the user to access the
      gossip on his behalf
- (optional) `proxy`: a proxy server checking the client's requests against the
  `authserver`

## Flow

TODO: update flow according to implementation

1. The user accesses the client through the browser.
    - `localhost:1234`
2. The user enters the name of the user whose gossip should be displayed (his name).
    - `localhost:1234/gossip?username=john`
3. The client tries to retrieve the gossip from the protected resource.
    - `localhost:8000/gossip/john`
4. The protected resource forwards the client to the authorization server.
    - `localhost:8443/authorizationForm?client_id=client&username=john`
    - a `forward` parameter to `localhost:1234/gossip?username=john` is also
      provided
5. The user enters the password and authorizes the client to access the gossip.
    - `client_id` and `username` from the request are stored in hidden fields
    - the `forward` parameter must also be stored in a hidden field
    - `localhost:8443/authorization` (form params: `client_id`, `username`, `password`, `forward`)
6. The auth server issues a `client_secret` and forwards the user back to the client.
    - `localhost:1234/gossip?username=john&client_secret=...`
7. The client now asks for a token for `username` with its `client_id`/`client_secret` credentials.
    - `localhost:8443/token?client_id=client&client_secret=[...]&username=john`
8. The auth server issues and stores an access token.
9. The client accesses the gossip resource on the user's behalf using the `access_token`:
    - `localhost:8000/gossip/john` with the `Authorization: Bearer ...` header
10. The protected resource checks the `access_token` against the auth server.
    - `localhost:8443/accesscheck?token=...`
11. The auth server validates the token and sends back status 200.
12. The protected resource sends the gossip to the client.
13. The client displays the gossip retrieved.
