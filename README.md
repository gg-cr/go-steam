# go-steam (gg-cr fork)

Go library for automating Steam network interactions. Forked from [paralin/go-steam](https://github.com/paralin/go-steam), which itself originates from [Philipp15b/go-steam](https://github.com/Philipp15b/go-steam) (a Go port of [SteamKit2](https://github.com/SteamRE/SteamKit)).

This fork adds **modern JWT-based authentication** (refresh tokens) required since Valve deprecated sentry files and login keys in mid-2023.

## What this fork changes

### Refresh token login (`LogOn` with `RefreshToken`)

Steam's CM now accepts JWT refresh tokens in the `access_token` field (108) of `CMsgClientLogon`. This fork makes that work:

```go
client.Auth.LogOn(&steam.LogOnDetails{
    Username:     "mybot",
    RefreshToken: savedRefreshToken, // JWT, ~200 day lifetime
})
```

**Technical detail:** `proto.Marshal` silently drops field 108 because `golang/protobuf` v1.5+ delegates to the v2 runtime which uses file descriptors, and field 108 isn't in the original proto descriptor. This fork adds `WriteRaw()` to manually append the wire-encoded field after marshaling.

### LogOnCredentials (CM-based auth flow)

A full `BeginAuthSessionViaCredentials` implementation over the CM wire protocol is included (`auth.go`). It handles RSA key exchange, Steam Guard (email/device/TOTP), machine trust tokens, and token polling. However, **this flow is unreliable** with paralin's wire protocol — `EMsg_ServiceMethodCallFromClientNonAuthed` causes EOF disconnects on some CM servers. 

**Recommended approach:** Use Steam's HTTP API (`IAuthenticationService/*`) to obtain the refresh token, then pass it to `LogOn`. See [ggcr-sentry](https://github.com/gg-cr/platform/tree/main/services/ggcr-sentry) for a working implementation.

### Other additions

- `Authenticator` interface for pluggable Steam Guard code input
- `Auth.Details` exported for token persistence between sessions
- `GuardData` support in `LogOnDetails` (machine trust, skips email code on re-auth)
- `JobHandlers` map + `GetNextJobId()` for async service method RPC routing
- `EMsg_ServiceMethodCallFromClientNonAuthed` (9804) and `EMsg_ClientHello` (9805)
- Full `CAuthentication_*` protobuf types in `protocol/protobuf/auth.pb.go`
- `ESessionPersistence` enum with proper proto runtime registration
- `AccessToken` field (108) added to `CMsgClientLogon`

## Installation

```
go get github.com/gg-cr/go-steam
```

Use with our [go-dota2 fork](https://github.com/gg-cr/go-dota2) which imports this module directly.

## Protobuf conflict with go-dota2

Both go-steam and go-dota2 register protobuf extension field 50000 (`MessageOptions`). If you use both, set this **before any proto packages are imported**:

```go
func init() {
    os.Setenv("GOLANG_PROTOBUF_REGISTRATION_CONFLICT", "ignore")
}
```

Put this in a package imported via blank import (`_ "yourpkg/protofix"`) as the first import in `main.go`.

## Upstream

- Original: [Philipp15b/go-steam](https://github.com/Philipp15b/go-steam)
- Paralin fork: [paralin/go-steam](https://github.com/paralin/go-steam) (upstream we track)
- Modern auth reference: [0xAozora/go-steam](https://github.com/0xAozora/go-steam)
- Our Dota 2 GC fork: [gg-cr/go-dota2](https://github.com/gg-cr/go-dota2)

## License

BSD 3-Clause. See [LICENSE.txt](LICENSE.txt).
