# nsqlitehttp

An HTTP client package for directly interacting with the **NSQLite database
engine** over HTTP/HTTPS.

Use this if you want a lower-level approach than the
[`database/sql`](https://pkg.go.dev/database/sql) interface provided by
[`nsqlitego`](https://github.com/nsqlite/nsqlitego).

## Features

- Issue raw JSON-based requests to an NSQLite server.
- Customizable headers, request methods, and endpoints.
- Handle responses as JSON or plain text.

## Installation

```bash
go get github.com/nsqlite/nsqlitego
```

> **Note**: This package is part of the
> [`nsqlitego`](https://github.com/nsqlite/nsqlitego) repository.\
> Import it as:
>
> ```go
> import "github.com/nsqlite/nsqlitego/nsqlitehttp"
> ```

## Usage

### Creating a Client

```go
client, err := nsqlitehttp.NewClient("http://localhost:9876?authToken=myToken")
if err != nil {
  panic(err)
}
```

### Sending Queries

```go
resp, err := client.Query(nsqlitehttp.Query{
  Query:  "SELECT id, name FROM users WHERE id > ?",
  Params: []any{100},
  // TxId can be optionally set if you are managing transactions at this level
})
if err != nil {
  panic(err)
}

fmt.Printf("Response Type: %s\n", resp.Type)
if resp.Type == nsqlitehttp.QueryResponseRead {
  // Access resp.Columns, resp.Values, etc.
}
```

### Ping / Health Check

```go
if err := client.Ping(); err != nil {
  fmt.Println("Server is not healthy:", err)
} else {
  fmt.Println("Server is up and running.")
}
```
