<p align="center">
  <p align="center">
    <img align="center" width="300" src="https://raw.githubusercontent.com/nsqlite/nsqlite/main/assets/NSQLite.png"/>
  </p>
  <p align="center">
    SQLite Over The Network
  </p>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/nsqlite/nsqlitego">
    <img src="https://pkg.go.dev/badge/github.com/nsqlite/nsqlitego" alt="Go Reference"/>
  </a>
  <a href="https://goreportcard.com/report/nsqlite/nsqlitego">
    <img src="https://goreportcard.com/badge/nsqlite/nsqlitego" alt="Go Report Card"/>
  </a>
  <a href="https://github.com/nsqlite/nsqlitego/releases/latest">
    <img src="https://img.shields.io/github/release/nsqlite/nsqlitego.svg" alt="Release Version"/>
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/nsqlite/nsqlitego.svg" alt="License"/>
  </a>
  <a href="https://github.com/nsqlite/nsqlitego">
    <img src="https://img.shields.io/github/stars/nsqlite/nsqlitego?style=flat&label=github+stars"/>
  </a>
</p>

# nsqlitego

NSQLite Go Driver, a driver for the
[**NSQLite database engine**](https://github.com/nsqlite/nsqlite) — compatible
with the standard [`database/sql`](https://pkg.go.dev/database/sql) package.

## Features

- Communicates with the **NSQLite database engine** over HTTP/HTTPS.
- Implements `database/sql/driver` interfaces for seamless integration.
- Supports transactions, prepared statements, and custom DSN parsing.
- Zero dependencies outside the standard library.

## Installation

```bash
go get github.com/nsqlite/nsqlitego
```

Ensure that you are using Go modules (`go mod init`) in your project.

## Usage

### Basic Example

Below is a concise example showing how to open a database and execute a simple
query:

```go
import (
  "database/sql"
  "fmt"
  _ "github.com/nsqlite/nsqlitego"
)

func main() {
  db, err := sql.Open("nsqlite", "http://localhost:9876?authToken=secret")
  if err != nil {
    panic(err)
  }
  defer db.Close()

  if err := db.Ping(); err != nil {
    panic("error pinging database: " + err.Error())
  }

  rows, err := db.Query("SELECT id, name FROM users")
  if err != nil {
    panic(err)
  }
  defer rows.Close()

  for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
      panic(err)
    }
    fmt.Println(id, name)
  }
}
```

### Transactions

Transactions are straightforward; they follow the standard pattern in
`database/sql`:

```go
tx, _ := db.Begin()
defer tx.Rollback()

_, _ = tx.Exec("INSERT INTO users(name) VALUES(?)", "Alice")
_, _ = tx.Exec("INSERT INTO users(name) VALUES(?)", "Bob")
_, _ = tx.Exec("INSERT INTO users(name) VALUES(?)", "Charlie")

if err := tx.Commit(); err != nil {
  // ...
}
```

Errors are ignored for brevity, but you should always handle them in your code.

## Additional Packages

These packages are included in this repository, so no additional installation is required.

- **[nsqlitedsn](nsqlitedsn/README.md)** – Provides convenient parsing
  and manipulation of NSQLite connection strings.
- **[nsqlitehttp](nsqlitehttp/README.md)** – An alternative way to
  access the **NSQLite database engine** directly over HTTP, offering more
  granular control than the `database/sql` layer.

## License

This project is licensed under the MIT license. See [LICENSE](LICENSE) for details.
