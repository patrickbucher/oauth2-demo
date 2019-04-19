# oauth2-demo

toy applications to demonstrate OAuth 2

components:

- `resource`: the application holding and serving the protected resources (some random gossip)
    - port 8000
    - endpoint `/gossip/[username]`: receive the gossip from a given user
- `authserver`: the server that issues and checks access tokens
    - port 8443
    - endpoint `/authorization?client=[client]&username=[username]&password=[password]`: authenticate a user and grant access rights to the client, return a `client_secret`
    - endpoint `/token?client_id=[client_id]&client_secret=[client_secret]&username=[username]`: issue a new `access_token` for an authorized client for the username specified
    - endpoint `/accesscheck?acces_token=[access_token]`: checks if a given `access_token` is (still) valid
- `client`: the client application that requests access to the protected resources
    - port 1234 
    - endpoint `/index.html`: web frontend that asks the user to access the gossip on his behalf
- (optional) `proxy`: a proxy server checking the client's requests against the `authserver`
