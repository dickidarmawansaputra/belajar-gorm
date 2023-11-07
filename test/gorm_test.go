package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dickidarmawansaputra/belajar-gorm/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func OpenConnection() *gorm.DB {
	dsn := "dickids:rahasia@tcp(127.0.0.1:3306)/belajar-gorm?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),

		// tips performa GORM
		// SkipDefaultTransaction: true,
		// PrepareStmt: true,
		// Gunakan select agar tidak ambil semua kolomnya
		// jika gunakan query find dg large query baiknya gunakan lazy result menggunakan Rows()
		// table split
	})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db
}

var db = OpenConnection()

func TestOpenConnection(t *testing.T) {
	assert.NotNil(t, db)
}

func TestExecuteSql(t *testing.T) {
	err := db.Exec("insert into sample(id, name) values(?, ?)", "1", "dicki").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(id, name) values(?, ?)", "2", "a").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(id, name) values(?, ?)", "3", "b").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(id, name) values(?, ?)", "4", "c").Error
	assert.Nil(t, err)

}

type Sample struct {
	Id   string
	Name string
}

func TestRawSql(t *testing.T) {
	var sample Sample
	err := db.Raw("SELECT id, name FROM sample WHERE id = ?", "1").Scan(&sample).Error
	assert.Nil(t, err)
	assert.Equal(t, "dicki", sample.Name)

	var samples []Sample
	err = db.Raw("SELECT * FROM sample").Scan(&samples).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(samples))
}

func TestSqlRaw(t *testing.T) {
	rows, err := db.Raw("SELECT * FROM sample").Rows()
	assert.Nil(t, err)
	defer rows.Close()

	var samples []Sample
	for rows.Next() {
		var id string
		var name string

		err := rows.Scan(&id, &name)
		assert.Nil(t, err)

		samples = append(samples, Sample{
			Id:   id,
			Name: name,
		})
	}
	assert.Equal(t, 4, len(samples))
}

func TestSqlRawGorm(t *testing.T) {
	rows, err := db.Raw("SELECT * FROM sample").Rows()
	assert.Nil(t, err)
	defer rows.Close()

	var samples []Sample
	for rows.Next() {
		// gunakan scanrows bawaan GORM
		err := db.ScanRows(rows, &samples)
		assert.Nil(t, err)
	}
	assert.Equal(t, 4, len(samples))
}

func TestCreateUser(t *testing.T) {
	user := model.User{
		Id:       "1",
		Password: "rahasia",
		Name: model.Name{
			FirstName:  "Dicki",
			MiddleName: "Darmawan",
			LastName:   "Saputra",
		},
		Information: "ini akan di ignore",
	}

	response := db.Create(&user)
	assert.Nil(t, response.Error)
	assert.Equal(t, int64(1), response.RowsAffected)
}

func TestBatchInsert(t *testing.T) {
	var users []model.User
	for i := 2; i < 10; i++ {
		users = append(users, model.User{
			Id:       strconv.Itoa(i),
			Password: "rahasia",
			Name: model.Name{
				FirstName: "User " + strconv.Itoa(i),
			},
		})
	}

	result := db.Create(&users)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(8), result.RowsAffected)
}

func TestDatabaseTransaction(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&model.User{Id: "13", Password: "rahasia", Name: model.Name{FirstName: "User 13"}}).Error
		if err != nil {
			return err
		}

		err = tx.Create(&model.User{Id: "11", Password: "rahasia", Name: model.Name{FirstName: "User 11"}}).Error
		if err != nil {
			return err
		}

		return nil
	})
	assert.Nil(t, err)
}

// tapi tidak disarankan manual seperti ini
func TestDatabaseManualTransaction(t *testing.T) {
	tx := db.Begin()
	defer tx.Rollback()

	err := tx.Create(&model.User{Id: "15", Password: "rahasia", Name: model.Name{FirstName: "User 15"}}).Error
	assert.Nil(t, err)

	err = tx.Create(&model.User{Id: "14", Password: "rahasia", Name: model.Name{FirstName: "User 14"}}).Error
	assert.Nil(t, err)

	if err == nil {
		tx.Commit()
	}
}

func TestQuerySingleObject(t *testing.T) {
	user := model.User{}
	err := db.First(&user).Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.Id)

	user = model.User{}
	err = db.Last(&user).Error
	assert.Nil(t, err)
	assert.Equal(t, "9", user.Id)

	// kalo ambil 1 data prefer pake take aja, soalnya pengurutan di first juga jadi tidak berguna karna by id hanya 1 data aja
	user = model.User{}
	err = db.Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
}

func TestQuerySingleObjectInlineCondition(t *testing.T) {
	user := model.User{}
	// ada order by
	err := db.First(&user, "id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.Id)

	user = model.User{}
	err = db.Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.Id)
}

func TestQueryAllObject(t *testing.T) {
	var users []model.User
	err := db.Find(&users, "id in ?", []string{"1", "2", "3", "4"}).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(users))
}

func TestQueryCondition(t *testing.T) {
	var users []model.User
	err := db.Where("first_name like ?", "%User%").
		Where("password = ?", "rahasia").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 13, len(users))
}

func TestQueryOrOperator(t *testing.T) {
	var users []model.User
	err := db.Where("first_name like ?", "%User%").
		Or("password = ?", "rahasia").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 14, len(users))
}

func TestQueryNotOperator(t *testing.T) {
	var users []model.User
	err := db.Not("first_name like ?", "%User%").
		Where("password = ?", "rahasia").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestQuerySelectFields(t *testing.T) {
	var users []model.User
	err := db.Select("first_name").Find(&users).Error
	assert.Nil(t, err)

	for _, user := range users {
		assert.NotNil(t, user.Id)
		assert.NotEqual(t, "", user.Name.FirstName)
	}
}

func TestQueryStructCondition(t *testing.T) {
	userCondition := model.User{
		Name: model.Name{
			FirstName: "User 5",
			LastName:  "", // tidak bisa, karna dianggap default value. gunakan map condition
		},
		Password: "rahasia",
	}

	var users []model.User
	err := db.Where(userCondition).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestQueryMapCondition(t *testing.T) {
	mapCondition := map[string]string{
		"last_name": "",
	}

	var users []model.User
	err := db.Where(mapCondition).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 13, len(users))
}

func TestOrderLimitOffset(t *testing.T) {
	var users []model.User
	err := db.Order("id asc, first_name desc").Limit(5).Offset(5).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
}

type UserResponse struct {
	ID        string
	FirstName string
	LastName  string
}

func TestQueryNonModel(t *testing.T) {
	var users []UserResponse

	err := db.Model(&model.User{}).Select("id", "first_name", "last_name").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 14, len(users))
}

func TestUpdate(t *testing.T) {
	user := model.User{}
	err := db.Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)

	user.Name.FirstName = "Dickids"
	// Save() secara default update semua kolom
	err = db.Save(&user).Error
	assert.Nil(t, err)
}

func TestUpdateSelectedColumns(t *testing.T) {
	err := db.Model(&model.User{}).Where("id = ?", "1").Updates(map[string]interface{}{
		"middle_name": "",
		"last_name":   "Update",
	}).Error
	assert.Nil(t, err)

	err = db.Model(&model.User{}).Where("id = ?", "1").Update("password", "rahasiailahi").Error
	assert.Nil(t, err)

	err = db.Where("id = ?", "1").Updates(model.User{
		Name: model.Name{
			FirstName:  "Dicki",
			MiddleName: "Darmawan",
			LastName:   "Saputra",
		},
		Password: "rahasia",
	}).Error
	assert.Nil(t, err)
}

func TestAutoIncrement(t *testing.T) {
	for i := 0; i < 10; i++ {
		userLog := model.UserLog{
			UserId: "1",
			Action: "test",
		}

		err := db.Create(&userLog).Error
		assert.Nil(t, err)
		assert.NotEqual(t, 0, userLog.ID)

		fmt.Println(userLog.ID)
	}
}

func TestSaveOrUpdate(t *testing.T) {
	userLog := model.UserLog{
		UserId: "1",
		Action: "test",
	}

	err := db.Save(&userLog).Error //create
	assert.Nil(t, err)

	userLog.UserId = "2"
	err = db.Save(&userLog).Error //update
	assert.Nil(t, err)
}

func TestSaveOrUpdateNonAutoIncrement(t *testing.T) {
	user := model.User{
		Id: "99",
		Name: model.Name{
			FirstName: "User 99",
		},
	}

	err := db.Save(&user).Error
	assert.Nil(t, err)

	user.Name.FirstName = "User 99 updated"
	err = db.Save(&user).Error
	assert.Nil(t, err)
}

func TestConflict(t *testing.T) {
	user := model.User{
		Id: "88",
		Name: model.Name{
			FirstName: "User 88",
		},
	}

	// buat seperti save untuk ngatasin conflict
	err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&user).Error
	assert.Nil(t, err)
}

func TestDelete(t *testing.T) {
	var user model.User
	err := db.Take(&user, "id = ?", "88").Error
	assert.Nil(t, err)

	// cara 1
	err = db.Delete(&user).Error
	assert.Nil(t, err)

	// cara 2
	err = db.Delete(&model.User{}, "id = ?", "99").Error
	assert.Nil(t, err)

	// cara 3
	err = db.Where("id = ?", "77").Delete(&model.User{}).Error
	assert.Nil(t, err)
}

func TestSoftDelete(t *testing.T) {
	todo := model.Todo{
		UserId:      "1",
		Title:       "todo 1",
		Description: "deskripsi todo 1",
	}

	err := db.Create(&todo).Error
	assert.Nil(t, err)

	err = db.Delete(&todo).Error
	assert.Nil(t, err)
	assert.NotNil(t, todo.DeletedAt)

	var todos []model.Todo
	err = db.Find(&todos).Error
	assert.Nil(t, err)
	assert.Equal(t, 0, len(todos))
}

func TestHardDelete(t *testing.T) {
	var todo model.Todo

	// gunakan unscoped() untuk mengambil yg softdelete juga
	err := db.Unscoped().First(&todo, "id = ?", "1").Error
	assert.Nil(t, err)

	err = db.Unscoped().Delete(&todo).Error
	assert.Nil(t, err)

	var todos []model.Todo
	err = db.Unscoped().Find(&todos).Error
	assert.Nil(t, err)
	assert.Equal(t, 0, len(todos))
}

func TestLocking(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&user, "id = ?", "1").Error
		if err != nil {
			return err
		}

		user.Name.FirstName = "Roy"
		user.Name.LastName = "Marten"
		err = tx.Save(&user).Error
		if err != nil {
			return err
		}

		return nil
	})
	assert.Nil(t, err)
}

func TestWallet(t *testing.T) {
	wallet := model.Wallet{
		Id:      "1",
		UserId:  "1",
		Balance: 1000,
	}

	err := db.Create(&wallet).Error
	assert.Nil(t, err)
}

// one to one relationship
func TestHasOne(t *testing.T) {
	var user model.User
	err := db.Model(&model.User{}).Preload("Wallet").Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.Id)
	assert.Equal(t, "1", user.Wallet.Id)
}

// Jika one to one sebaiknya gunakan join saja karna cukup sekali query
func TestHasOneJoin(t *testing.T) {
	var user model.User
	err := db.Model(&model.User{}).Joins("Wallet").Take(&user, "users.id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.Id)
	assert.Equal(t, "1", user.Wallet.Id)
}

// GORM akan auto create update jika ada relasi
func TestAutoCreateUpdate(t *testing.T) {
	user := model.User{
		Id: "200",
		Name: model.Name{
			FirstName: "User 200",
		},
		Password: "rahasia",
		Wallet: model.Wallet{
			Id:      "2",
			UserId:  "200",
			Balance: 100000,
		},
	}

	err := db.Create(&user).Error
	assert.Nil(t, err)
}

func TestSkipAutoCreateUpdate(t *testing.T) {
	user := model.User{
		Id: "300",
		Name: model.Name{
			FirstName: "User 300",
		},
		Password: "rahasia",
		Wallet: model.Wallet{
			Id:      "3",
			UserId:  "300",
			Balance: 100000,
		},
	}

	err := db.Omit(clause.Associations).Create(&user).Error
	assert.Nil(t, err)
}

// one to many relationship
func TestHasMany(t *testing.T) {
	user := model.User{
		Id: "201",
		Name: model.Name{
			FirstName: "User 201",
		},
		Password: "rahasia",
		Wallet: model.Wallet{
			Id:      "4",
			UserId:  "201",
			Balance: 100000,
		},
		Addresses: []model.Address{
			{
				UserId:  "201",
				Address: "A",
			},
			{
				UserId:  "201",
				Address: "B",
			},
		},
	}

	err := db.Create(&user).Error
	assert.Nil(t, err)
}

func TestPreloadHasMany(t *testing.T) {
	var user []model.User
	err := db.Model(&model.User{}).Preload("Wallet").Preload("Addresses").Find(&user).Error
	assert.Nil(t, err)
}

// many to one relationship
func TestBelongsTo(t *testing.T) {
	fmt.Println("preload")
	var addresses []model.Address
	err := db.Model(&model.Address{}).Preload("User").Find(&addresses).Error
	assert.Nil(t, err)

	fmt.Println("join")
	err = db.Model(&model.Address{}).Joins("User").Find(&addresses).Error
	assert.Nil(t, err)
}

// cyclic misal user butuh relasi wallet di walet butuh relasi user
func TestBelongsToWallet(t *testing.T) {
	fmt.Println("preload")
	var wallets []model.Wallet
	err := db.Model(&model.Wallet{}).Preload("User").Find(&wallets).Error
	assert.Nil(t, err)

	fmt.Println("join")
	err = db.Model(&model.Wallet{}).Joins("User").Find(&wallets).Error
	assert.Nil(t, err)
}

func TestCreateManyToMany(t *testing.T) {
	product := model.Product{
		ID:    "p1",
		Name:  "product 1",
		Price: 1000,
	}

	err := db.Create(&product).Error
	assert.Nil(t, err)

	err = db.Table("user_like_product").Create(map[string]interface{}{
		"user_id":    "1",
		"product_id": "p1",
	}).Error
	assert.Nil(t, err)
}

func TestPreloadManyToManyProduct(t *testing.T) {
	var product model.Product
	err := db.Preload("LikedByUsers").Take(&product, "id = ?", "p1").Error
	assert.Nil(t, err)
}

func TestPreloadManyToManyUser(t *testing.T) {
	var user model.User
	err := db.Preload("LikeProducts").Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
}

func TestAssociationFind(t *testing.T) {
	var product model.Product
	err := db.Take(&product, "id = ?", "p1").Error
	assert.Nil(t, err)

	var users []model.User
	err = db.Model(&product).Where("first_name LIKE ?", "%Roy%").Association("LikedByUsers").Find(&users)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

// menambah di relasi many to many
func TestAssociationAppend(t *testing.T) {
	var user model.User
	err := db.Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)

	var product model.Product
	err = db.Take(&product, "id = ?", "p1").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Append(&user)
	assert.Nil(t, err)
}

// idealnya pake transaction
// replace hanya cocok untuk relasi one to one / belongs to
func TestAssociationReplace(t *testing.T) {
	db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		err := tx.Take(&user, "id = ?", "2").Error
		assert.Nil(t, err)

		wallet := model.Wallet{
			Id:      "w1",
			UserId:  user.Id,
			Balance: 5000,
		}

		err = tx.Model(&user).Association("Wallet").Replace(&wallet)
		assert.Nil(t, err)

		return nil
	})
}

func TestAssociationDelete(t *testing.T) {
	var user model.User
	err := db.Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)

	var product model.Product
	err = db.Take(&product, "id = ?", "p1").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Delete(&user)
	assert.Nil(t, err)
}

func TestAssociationClear(t *testing.T) {
	var product model.Product
	err := db.Take(&product, "id = ?", "p1").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Clear()
	assert.Nil(t, err)
}

func TestPreloadWithCondition(t *testing.T) {
	var user model.User
	err := db.Preload("Wallet", "balance > ?", 1000).Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
}

func TestNestedPreload(t *testing.T) {
	var wallet model.Wallet
	err := db.Preload("User.Addresses").Take(&wallet, "id = ?", "4").Error
	assert.Nil(t, err)

	// karna pointer di user jd di print semua
	fmt.Println(wallet)
	fmt.Println(wallet.User)
	fmt.Println(wallet.User.Addresses)
}

// preload all tidak akan load nested relation
func TestPreloadAll(t *testing.T) {
	var user model.User
	err := db.Preload(clause.Associations).Take(&user, "id = ?", "201").Error
	assert.Nil(t, err)
}

func TestJoinQuery(t *testing.T) {
	var users []model.User
	err := db.Joins("join wallets on wallets.user_id = users.id").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))

	err = db.Joins("Wallet").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 18, len(users))
}

func TestJoinQueryCondition(t *testing.T) {
	var users []model.User
	err := db.Joins("JOIN wallets ON wallets.user_id = users.id AND wallets.balance > ?", 1000).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(users))

	err = db.Joins("Wallet").Where("balance > ?", 1000).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(users))
}

func TestCountAggregation(t *testing.T) {
	var count int64
	err := db.Model(&model.User{}).Joins("Wallet").Where("balance > ?", 1000).Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, int64(4), count)
}

type AggregationResult struct {
	TotalBalance int64
	MinBalance   int64
	MaxBalance   int64
	AvgBalance   float64
}

func TestAggregation(t *testing.T) {
	var result AggregationResult
	err := db.Model(&model.Wallet{}).Select("sum(balance) as total_balance", "min(balance) as min_balance", "max(balance) as max_balance", "avg(balance) as avg_balance").Take(&result).Error
	assert.Nil(t, err)
	assert.Equal(t, int64(306000), result.TotalBalance)
	assert.Equal(t, int64(1000), result.MinBalance)
	assert.Equal(t, int64(100000), result.MaxBalance)
	assert.Equal(t, float64(61200), result.AvgBalance)
}

func TestAggregationGroupByHaving(t *testing.T) {
	var result []AggregationResult
	err := db.Model(&model.Wallet{}).Select("sum(balance) as total_balance", "min(balance) as min_balance", "max(balance) as max_balance", "avg(balance) as avg_balance").Joins("User").Group("user_id").Having("sum(balance) > ?", 1000).Find(&result).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(result))
}

// idealnya pake context
func TestGormWithContext(t *testing.T) {
	ctx := context.Background()

	var users []model.User
	err := db.WithContext(ctx).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 18, len(users))
}

func BrokeWalletBalance(db *gorm.DB) *gorm.DB {
	return db.Where("balance = ?", 0)
}

func SultanWalletBalance(db *gorm.DB) *gorm.DB {
	return db.Where("balance > ?", 5000)
}

func TestScopes(t *testing.T) {
	var wallets []model.Wallet
	err := db.Scopes(BrokeWalletBalance).Find(&wallets).Error
	assert.Nil(t, err)

	err = db.Scopes(SultanWalletBalance).Find(&wallets).Error
	assert.Nil(t, err)
}

// tetap disarankan pake migration yg support versioning
// bawaan gorm hanya digunakan untuk test di local
func TestMigrator(t *testing.T) {
	err := db.Migrator().AutoMigrate(&model.GuestBook{})
	assert.Nil(t, err)
}

func TestHooks(t *testing.T) {
	user := model.User{
		Password: "rahasia",
		Name: model.Name{
			FirstName: "Hooks",
		},
	}

	err := db.Create(&user).Error
	assert.Nil(t, err)
	assert.NotEqual(t, "", user.Id)

	fmt.Println(user.Id)
}
