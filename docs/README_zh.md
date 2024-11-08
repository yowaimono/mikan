# MinDB: 一个极简的内存数据库

MinDB 是一个轻量级的内存数据库，设计简单易用。它提供了基本的 CRUD 操作、查询功能，并支持唯一约束和默认值。MinDB 非常适合小型项目、原型开发或任何需要快速易用的数据库场景，而无需使用功能齐全的 SQL 数据库。

## 功能特性

- **CRUD 操作**：执行基本的创建、读取、更新和删除操作。
- **查询功能**：使用简单的查询接口过滤、排序和限制结果。
- **唯一约束**：在特定字段上定义唯一约束，以确保数据完整性。
- **默认值**：为字段设置默认值，简化数据输入。
- **预写日志 (WAL)**：将数据更改持久化到日志文件中，以便恢复。
- **自动接口生成**：自动为您的模型生成 `Keyable` 接口，减少样板代码。

## 安装

要在您的 Go 项目中使用 MinDB，只需导入它：

```go
import "github.com/yourusername/minfs"
```

## 快速开始

### 1. 初始化数据库

首先，初始化 MinDB 实例：

```go
db := minfs.GetDB()
```

### 2. 注册您的模型

将您的模型注册到数据库中。MinDB 将自动为您的模型生成必要的接口：

```go
type User struct {
    ID   string `min:"unique"`
    Name string `min:"unique"`
    Age  int    `min:"default 18"`
}

db.AutoCreate(&User{})
```

### 3. 执行 CRUD 操作

#### 添加数据

```go
user := User{Name: "Alice", Age: 25}
if err := db.Add(&user); err != nil {
    log.Fatalf("添加用户错误 %v", err)
}
```

#### 检索数据

```go
var users []User
err := db.Query(&User{}).Find().Unwrap(&users)
if err != nil {
    log.Fatalf("查询错误 %v", err)
}
fmt.Println("用户列表:", users)
```

#### 更新数据

```go
user.Age = 26
if err := db.Up(&user); err != nil {
    log.Fatalf("更新用户错误 %v", err)
}
```

#### 删除数据

```go
if err := db.Del(&user); err != nil {
    log.Fatalf("删除用户错误 %v", err)
}
```

### 4. 查询数据

MinDB 支持多种查询操作：

#### 过滤数据

```go
var filteredUsers []User
err := db.Query(&User{}).Where("Age", 25).Find().Unwrap(&filteredUsers)
if err != nil {
    log.Fatalf("过滤错误 %v", err)
}
fmt.Println("过滤后的用户列表:", filteredUsers)
```

#### 排序数据

```go
var sortedUsers []User
err := db.Query(&User{}).OrderBy("Age", true).Find().Unwrap(&sortedUsers)
if err != nil {
    log.Fatalf("排序错误 %v", err)
}
fmt.Println("排序后的用户列表:", sortedUsers)
```

#### 限制结果

```go
var limitedUsers []User
err := db.Query(&User{}).Limit(3).Find().Unwrap(&limitedUsers)
if err != nil {
    log.Fatalf("限制错误 %v", err)
}
fmt.Println("限制后的用户列表:", limitedUsers)
```

### 5. 关闭数据库

完成后，关闭数据库以释放资源：

```go
if err := db.Close(); err != nil {
    log.Fatalf("关闭数据库错误 %v", err)
}
```

## 高级功能

### 唯一约束

在字段上定义唯一约束，以确保没有两个记录在该字段上有相同的值：

```go
type User struct {
    ID   string `min:"unique"`
    Name string `min:"unique"`
    Age  int    `min:"default 18"`
}
```

### 默认值

为字段设置默认值，简化数据输入：

```go
type User struct {
    ID   string `min:"unique"`
    Name string `min:"unique"`
    Age  int    `min:"default 18"`
}
```

### 预写日志 (WAL)

MinDB 使用预写日志将数据更改持久化到日志文件中。这确保了在崩溃情况下可以恢复数据：

```go
table, err := db.getOrCreateTable("users")
if err != nil {
    log.Fatalf("创建表错误 %v", err)
}
table.writeWAL("ADD", user.String())
```

## 贡献

欢迎贡献！请随时打开问题或提交拉取请求。

## 许可证

MinDB 使用 MIT 许可证。有关更多详细信息，请参阅 [LICENSE](LICENSE) 文件。

---

MinDB 是一个简单但功能强大的工具，用于管理您的 Go 应用程序中的数据。无论您是构建小型原型还是大型项目，MinDB 都提供了您所需的灵活性和易用性。试试看，看看它如何简化您的数据管理任务！