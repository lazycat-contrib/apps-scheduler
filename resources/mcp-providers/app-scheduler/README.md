# App Scheduler MCP Provider

Endpoint:

```text
/mcp
```

LazyCat app-to-app callers should connect through:

```text
http://app.community.lazycat.app.czyt.apps-scheduler.lzcx/mcp
```

When calling from another LazyCat app, declare `lzcapp.user_delegate` in the caller package, obtain the current request's `X-HC-USER-TICKET`, and forward it with `X-HC-USER-ID` to this MCP endpoint.

Non-LazyCat callers can create a Bearer token from the App Scheduler settings page and pass:

```text
Authorization: Bearer <token>
```
