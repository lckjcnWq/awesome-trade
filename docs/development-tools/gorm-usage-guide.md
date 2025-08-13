# GORM框架详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [数据库连接](#数据库连接)
3. [模型定义](#模型定义)
4. [CRUD操作](#CRUD操作)
5. [查询操作](#查询操作)
6. [关联关系](#关联关系)
7. [事务处理](#事务处理)
8. [迁移和索引](#迁移和索引)
9. [高级特性](#高级特性)

## 基础概念

### 1.1 GORM简介

GORM是Go语言的ORM库，功能全面，开发效率高，支持关系型数据库操作。

```go
// 安装GORM
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql    // MySQL驱动
go get -u gorm.io/driver/postgres // PostgreSQL驱动
go get -u gorm.io/driver/sqlite   // SQLite驱动
```

### 1.2 基本使用

```go
package main

import (
    "gorm.io/gorm"
    "gorm.io/driver/mysql"
)

func main() {
    // 连接数据库
    dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // 自动迁移模式
    db.AutoMigrate(&User{})
}
```

## 数据库连接

### 2.1 MySQL连接

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func ConnectMySQL() (*gorm.DB, error) {
    dsn := "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info), // 启用SQL日志
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "t_",  // 表名前缀
            SingularTable: true,  // 使用单数表名
        },
    })
    
    if err != nil {
        return nil, err
    }

    // 获取通用数据库对象sql.DB，然后使用其提供的功能
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // 设置最大打开连接数
    sqlDB.SetMaxOpenConns(100)
    // 设置最大空闲连接数
    sqlDB.SetMaxIdleConns(10)
    // 设置连接的最大生存时间
    sqlDB.SetConnMaxLifetime(time.Hour)

    return db, nil
}
```

### 2.2 PostgreSQL连接

```go
import "gorm.io/driver/postgres"

func ConnectPostgreSQL() (*gorm.DB, error) {
    dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    return db, err
}
```

### 2.3 SQLite连接

```go
import "gorm.io/driver/sqlite"

func ConnectSQLite() (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    return db, err
}
```

## 模型定义

### 3.1 基本模型

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Name      string         `gorm:"size:100;not null"`
    Email     string         `gorm:"uniqueIndex;size:100"`
    Age       int            `gorm:"check:age > 0"`
    Birthday  *time.Time
    Active    bool           `gorm:"default:true"`
}
```

### 3.2 字段标签

```go
type Product struct {
    ID          uint      `gorm:"primaryKey;autoIncrement"`
    Code        string    `gorm:"uniqueIndex;size:50;not null"`
    Name        string    `gorm:"size:100;not null;comment:产品名称"`
    Price       float64   `gorm:"precision:10;scale:2;not null"`
    Description string    `gorm:"type:text"`
    CategoryID  uint      `gorm:"index"`
    CreatedAt   time.Time `gorm:"autoCreateTime"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// 自定义表名
func (Product) TableName() string {
    return "products"
}
```

### 3.3 嵌入结构体

```go
type BaseModel struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
    BaseModel
    Name  string `gorm:"size:100"`
    Email string `gorm:"uniqueIndex"`
}
```

## CRUD操作

### 4.1 创建记录

```go
// 创建单条记录
user := User{Name: "张三", Email: "zhangsan@example.com", Age: 25}
result := db.Create(&user)
if result.Error != nil {
    // 处理错误
    log.Printf("创建用户失败: %v", result.Error)
}
fmt.Printf("创建的用户ID: %d, 影响行数: %d", user.ID, result.RowsAffected)

// 批量创建
users := []User{
    {Name: "李四", Email: "lisi@example.com"},
    {Name: "王五", Email: "wangwu@example.com"},
}
db.Create(&users)

// 分批创建（避免内存问题）
db.CreateInBatches(users, 100) // 每批100条
```

### 4.2 查询记录

```go
// 查询单条记录
var user User
db.First(&user, 1)                 // 根据主键查询
db.First(&user, "name = ?", "张三") // 根据条件查询
db.Take(&user)                     // 获取一条记录，没有指定排序字段
db.Last(&user)                     // 获取最后一条记录

// 查询多条记录
var users []User
db.Find(&users)                              // 查询所有用户
db.Where("age > ?", 18).Find(&users)        // 条件查询
db.Where("name IN ?", []string{"张三", "李四"}).Find(&users)
```

### 4.3 更新记录

```go
// 更新单个字段
db.Model(&user).Update("name", "新名字")

// 更新多个字段
db.Model(&user).Updates(User{Name: "新名字", Age: 30})
db.Model(&user).Updates(map[string]interface{}{"name": "新名字", "age": 30})

// 批量更新
db.Model(&User{}).Where("age > ?", 18).Updates(User{Active: true})

// 更新选定字段
db.Model(&user).Select("name", "age").Updates(User{Name: "新名字", Age: 30})

// 忽略某些字段
db.Model(&user).Omit("created_at").Updates(User{Name: "新名字"})
```

### 4.4 删除记录

```go
// 软删除（如果模型有DeletedAt字段）
db.Delete(&user, 1)
db.Where("age < ?", 18).Delete(&User{})

// 永久删除
db.Unscoped().Delete(&user, 1)

// 物理删除（没有DeletedAt字段的模型）
db.Delete(&Product{}, 1)
```

## 查询操作

### 5.1 条件查询

```go
// Where条件
db.Where("name = ?", "张三").First(&user)
db.Where("name <> ?", "张三").Find(&users)
db.Where("name IN ?", []string{"张三", "李四"}).Find(&users)
db.Where("name LIKE ?", "%张%").Find(&users)
db.Where("age BETWEEN ? AND ?", 18, 65).Find(&users)

// 结构体条件
db.Where(&User{Name: "张三", Age: 25}).First(&user)

// Map条件
db.Where(map[string]interface{}{"name": "张三", "age": 25}).Find(&users)
```

### 5.2 排序和分页

```go
// 排序
db.Order("age desc, name").Find(&users)
db.Order("age desc").Order("name").Find(&users)

// 分页
db.Limit(10).Offset(20).Find(&users)

// 计数
var count int64
db.Model(&User{}).Where("age > ?", 18).Count(&count)
```

### 5.3 分组和聚合

```go
type Result struct {
    Age   int
    Count int64
}

var results []Result
db.Model(&User{}).Select("age, count(*) as count").Group("age").Having("count > ?", 1).Find(&results)

// 聚合函数
var totalAge int64
db.Model(&User{}).Select("sum(age)").Row().Scan(&totalAge)
```

## 关联关系

### 6.1 一对一关系

```go
type User struct {
    gorm.Model
    Name    string
    Profile Profile
}

type Profile struct {
    gorm.Model
    UserID uint
    Bio    string
}

// 预加载
var user User
db.Preload("Profile").First(&user)

// 创建关联
user := User{Name: "张三"}
profile := Profile{Bio: "这是个人简介"}
user.Profile = profile
db.Create(&user)
```

### 6.2 一对多关系

```go
type User struct {
    gorm.Model
    Name   string
    Orders []Order
}

type Order struct {
    gorm.Model
    UserID uint
    Amount float64
}

// 预加载
var user User
db.Preload("Orders").First(&user)

// 条件预加载
db.Preload("Orders", "amount > ?", 100).First(&user)
```

### 6.3 多对多关系

```go
type User struct {
    gorm.Model
    Name  string
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    gorm.Model
    Name  string
    Users []User `gorm:"many2many:user_roles;"`
}

// 预加载
var user User
db.Preload("Roles").First(&user)

// 关联操作
db.Model(&user).Association("Roles").Append(&role)
db.Model(&user).Association("Roles").Replace(&roles)
db.Model(&user).Association("Roles").Delete(&role)
db.Model(&user).Association("Roles").Clear()
```

## 事务处理

### 7.1 手动事务

```go
func TransferMoney(fromUserID, toUserID uint, amount float64) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 扣除发送方余额
        if err := tx.Model(&User{}).Where("id = ?", fromUserID).
            Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }

        // 增加接收方余额
        if err := tx.Model(&User{}).Where("id = ?", toUserID).
            Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }

        // 创建转账记录
        transfer := Transfer{
            FromUserID: fromUserID,
            ToUserID:   toUserID,
            Amount:     amount,
        }
        return tx.Create(&transfer).Error
    })
}
```

### 7.2 手动控制事务

```go
func ManualTransaction() error {
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    if err := tx.Error; err != nil {
        return err
    }

    if err := tx.Create(&user1).Error; err != nil {
        tx.Rollback()
        return err
    }

    if err := tx.Create(&user2).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```

## 迁移和索引

### 8.1 自动迁移

```go
// 自动迁移
db.AutoMigrate(&User{}, &Product{}, &Order{})

// 检查表是否存在
if !db.Migrator().HasTable(&User{}) {
    db.Migrator().CreateTable(&User{})
}

// 检查列是否存在
if !db.Migrator().HasColumn(&User{}, "Name") {
    db.Migrator().AddColumn(&User{}, "Name")
}
```

### 8.2 索引管理

```go
// 创建索引
db.Migrator().CreateIndex(&User{}, "Name")
db.Migrator().CreateIndex(&User{}, "idx_user_name")

// 删除索引
db.Migrator().DropIndex(&User{}, "Name")

// 检查索引是否存在
db.Migrator().HasIndex(&User{}, "Name")
```

## 高级特性

### 9.1 钩子函数

```go
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
    // 创建前的处理
    if u.Name == "" {
        return errors.New("name cannot be empty")
    }
    return
}

func (u *User) AfterCreate(tx *gorm.DB) (err error) {
    // 创建后的处理
    log.Printf("User %s created with ID %d", u.Name, u.ID)
    return
}
```

### 9.2 自定义数据类型

```go
import (
    "database/sql/driver"
    "encoding/json"
)

type JSON json.RawMessage

func (j *JSON) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("type assertion to []byte failed")
    }
    *j = bytes
    return nil
}

func (j JSON) Value() (driver.Value, error) {
    if len(j) == 0 {
        return nil, nil
    }
    return []byte(j), nil
}

type User struct {
    ID       uint `gorm:"primaryKey"`
    Name     string
    Metadata JSON `gorm:"type:json"`
}
```

### 9.3 原生SQL

```go
// 原生查询
var users []User
db.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(&users)

// 执行原生SQL
db.Exec("UPDATE users SET active = ? WHERE age > ?", true, 18)

// 命名参数
db.Raw("SELECT * FROM users WHERE name = @name AND age > @age", 
    sql.Named("name", "张三"), sql.Named("age", 18)).Scan(&users)
```
