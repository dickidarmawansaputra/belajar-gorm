package model

import (
	"time"

	"gorm.io/gorm"
)

// check convention di GORM
type User struct {
	Id       string `gorm:"column:id;primaryKey;<-:create"`
	Password string `gorm:"column:password"`
	// GORM embbeded
	Name      Name      `gorm:"embedded"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create"`
	UpdatedAt time.Time `gorm:"column:created_at;autoCreateTime;autoUpdateTime"`
	// contoh penerapan field permission
	// lebih lengkap di file pdfnya
	// seperti tanda <-: -  dll
	Information  string    `gorm:"-"`
	Wallet       Wallet    `gorm:"foreignKey:user_id;references:id"`
	Addresses    []Address `gorm:"foreignKey:user_id;references:id"`
	LikeProducts []Product `gorm:"many2many:user_like_product;foreignKey:id;joinForeignKey:user_id;references:id;joinReferences:product_id"`
}

// jika ingin merubah nama table
func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(db *gorm.DB) error {
	if u.Id == "" {
		u.Id = "user-" + time.Now().Format("20060102150405")
	}
	return nil
}

type Name struct {
	FirstName  string `gorm:"column:first_name"`
	MiddleName string `gorm:"column:middle_name"`
	LastName   string `gorm:"column:last_name"`
}

type UserLog struct {
	ID     int    `gorm:"column:id;primaryKey;autoIncrement"`
	UserId string `gorm:"column:user_id"`
	Action string `gorm:"column:action"`
	// ubah dari timestamp ke epoch time
	CreatedAt int64 `gorm:"column:created_at;autoCreateTime:milli;<-:create"`
	UpdatedAt int64 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (u *UserLog) TableName() string {
	return "user_logs"
}

type Todo struct {
	// gunakan struct GORM Model sebagai pengganti id, created_at, updated_at, deleted_at
	// catatan idnya auto increment jika tidak, makan buat manual
	gorm.Model
	// ID          int64          `gorm:"column:id;primaryKey;autoIncrement"`
	UserId      string `gorm:"column:user_id"`
	Title       string `gorm:"column:title"`
	Description string `gorm:"column:description"`
	// CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt   time.Time      `gorm:"column:created_at;autoCreateTime;autoUpdateTime"`
	// DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (u *Todo) TableName() string {
	return "todos"
}

type Wallet struct {
	Id        string         `gorm:"column:id"`
	UserId    string         `gorm:"column:user_id"`
	Balance   int64          `gorm:"column:balance"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:created_at;autoCreateTime;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
	// jika terjadi cyclic gunakan pointer
	// belongs to juga bisa jadi has one
	User *User `gorm:"foreignKey:user_id;references:id"`
}

func (u *Wallet) TableName() string {
	return "wallets"
}

type Address struct {
	gorm.Model
	UserId  string `gorm:"column:user_id"`
	Address string `gorm:"column:address"`
	User    User   `gorm:"foreignKey:user_id;references:id"`
}

func (u *Address) TableName() string {
	return "addresses"
}

type Product struct {
	ID           string    `gorm:"column:id;primaryKey"`
	Name         string    `gorm:"column:name"`
	Price        int64     `gorm:"column:price"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:created_at;autoCreateTime;autoUpdateTime"`
	LikedByUsers []User    `gorm:"many2many:user_like_product;foreignKey:id;joinForeignKey:product_id;references:id;joinReferences:user_id"`
}

func (u *Product) TableName() string {
	return "products"
}

type GuestBook struct {
	gorm.Model
	Name    string `gorm:"column:name"`
	Email   string `gorm:"column:email"`
	Message string `gorm:"column:message"`
}

func (u *GuestBook) TableName() string {
	return "guest_books"
}
