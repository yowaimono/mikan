                                           [简体中文]()

# MinDB: A Minimalistic In-Memory Database

MinDB is a lightweight, in-memory database designed for simplicity and ease of use. It provides basic CRUD operations, query capabilities, and support for unique constraints and default values. MinDB is ideal for small projects, prototypes, or any scenario where you need a quick and easy-to-use database without the overhead of a full-fledged SQL database.

## Features

- **CRUD Operations**: Perform basic Create, Read, Update, and Delete operations.
- **Query Capabilities**: Filter, sort, and limit results using a simple query interface.
- **Unique Constraints**: Define unique constraints on specific fields to ensure data integrity.
- **Default Values**: Set default values for fields to simplify data entry.
- **WAL (Write-Ahead Logging)**: Persist data changes to a log file for recovery purposes.
- **Automatic Interface Generation**: Automatically generate the `Keyable` interface for your models, reducing boilerplate code.

## Installation

To use MinDB in your Go project, simply import it:

```go
import "github.com/yourusername/minfs"
```

## Quick Start

### 1. Initialize the Database

First, initialize the MinDB instance:

```go
db := minfs.GetDB()
```

### 2. Register Your Models

Register your models with the database. MinDB will automatically generate the necessary interfaces for your models:

```go
type User struct {
    ID   string `min:"unique"`
    Name string `min:"unique"`
    Age  int    `min:"default 18"`
}

db.AutoCreate(&User{})
```

### 3. Perform CRUD Operations

#### Add Data

```go
user := User{Name: "Alice", Age: 25}
if err := db.Add(&user); err != nil {
    log.Fatalf("add user error %v", err)
}
```

#### Retrieve Data

```go
var users []User
err := db.Query(&User{}).Find().Unwrap(&users)
if err != nil {
    log.Fatalf("query error %v", err)
}
fmt.Println("Users:", users)
```

#### Update Data

```go
user.Age = 26
if err := db.Up(&user); err != nil {
    log.Fatalf("update user error %v", err)
}
```

#### Delete Data

```go
if err := db.Del(&user); err != nil {
    log.Fatalf("delete user error %v", err)
}
```

### 4. Query Data

MinDB supports a variety of query operations:

#### Filter Data

```go
var filteredUsers []User
err := db.Query(&User{}).Where("Age", 25).Find().Unwrap(&filteredUsers)
if err != nil {
    log.Fatalf("filter error %v", err)
}
fmt.Println("Filtered Users:", filteredUsers)
```

#### Sort Data

```go
var sortedUsers []User
err := db.Query(&User{}).OrderBy("Age", true).Find().Unwrap(&sortedUsers)
if err != nil {
    log.Fatalf("sort error %v", err)
}
fmt.Println("Sorted Users:", sortedUsers)
```

#### Limit Results

```go
var limitedUsers []User
err := db.Query(&User{}).Limit(3).Find().Unwrap(&limitedUsers)
if err != nil {
    log.Fatalf("limit error %v", err)
}
fmt.Println("Limited Users:", limitedUsers)
```

### 5. Close the Database

When you're done, close the database to release resources:

```go
if err := db.Close(); err != nil {
    log.Fatalf("close db error %v", err)
}
```

## Advanced Features

### Unique Constraints

Define unique constraints on fields to ensure that no two records have the same value for that field:

```go
type User struct {
    ID   string `min:"unique"`
    Name string `min:"unique"`
    Age  int    `min:"default 18"`
}
```

### Default Values

Set default values for fields to simplify data entry:

```go
type User struct {
    ID   string `min:"unique"`
    Name string `min:"unique"`
    Age  int    `min:"default 18"`
}
```

### Write-Ahead Logging (WAL)

MinDB uses Write-Ahead Logging to persist data changes to a log file. This ensures that data can be recovered in case of a crash:

```go
table, err := db.getOrCreateTable("users")
if err != nil {
    log.Fatalf("create table error %v", err)
}
table.writeWAL("ADD", user.String())
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

MinDB is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

MinDB is a simple, yet powerful tool for managing data in your Go applications. Whether you're building a small prototype or a larger project, MinDB provides the flexibility and ease of use you need. Give it a try and see how it can simplify your data management tasks!