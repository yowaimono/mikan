package main

import (
	"encoding/json"
	"mikan/db"
	"mikan/pkg/logger"
	"time"
)

type User struct {
	ID   string `min:"unique"`
	Name string `min:"unique"`
	Age  int    `min:"default 18"`
}

func (u *User) GetKey() string {
	return u.ID
}

func (u *User) SetKey(key string) {
	u.ID = key
}

// func (u *User) String() string {
// 	return fmt.Sprintf("%s,%s,%d", u.ID, u.Name, u.Age)
// }

// func (u *User) From(s string) error {
// 	parts := strings.Split(s, ",")
// 	if len(parts) != 3 {
// 		return fmt.Errorf("invalid data format")
// 	}
// 	u.ID = parts[0]
// 	u.Name = parts[1]
// 	age, err := strconv.Atoi(parts[2])
// 	if err != nil {
// 		return err
// 	}
// 	u.Age = age
// 	return nil
// }

func (u *User) String() string {
	jb, err := json.Marshal(u)

	if err != nil {
		panic(err)
	}

	return string(jb)
}

func (u *User) From(s string) error {
	return json.Unmarshal([]byte(s), u)
}

func (u *User) AfterCreate(data db.Keyable) error {
	logger.Info("AfterAdd user %v", data)
	return nil
}

func main() {
	// db := db.GetDB(&db.Config{
	// 	LogLevel: logger.DEBUG,
	// })
	// defer db.Close()
	// // logger.GetLogger().SetLevel(logger.LogLevelDebug)
	// // 注册结构体类型
	// db.AutoCreate(&User{})

	// // 插入十条数据
	// users := []User{
	// 	{Name: "Alice", Age: 25},
	// 	{Name: "Bob", Age: 30},
	// 	{Name: "Charlie", Age: 35},
	// 	{Name: "David", Age: 40},
	// 	{Name: "Eve", Age: 45},
	// 	{Name: "Frank", Age: 50},
	// 	{Name: "Grace", Age: 55},
	// 	{Name: "Heidi", Age: 60},
	// 	{Name: "Ivan", Age: 65},
	// 	{Name: "Judy", Age: 70},
	// }

	// for _, user := range users {
	// 	if err := db.Create(&user); err != nil {
	// 		logger.Debug("add user error %v", err)
	// 	}
	// }

	s := time.Now()
	logger.GetLogger().SetOutputType(logger.FILE, "app.log")

	for i := 0; i < 1000; i++ {
		logger.Error("error %d ", i)
	}

	d := time.Since(s)

	logger.Info("time %v", d)
	// logger.Fatal("fatal")
	time.Sleep(60 * time.Second)
	// logger.GetLogger().Close()
}
