package model

import (
	"errors"
	"message-pusher/common"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func createRootAccountIfNeed() error {
	var user User
	err := DB.First(&user).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	common.SysLog("no user exists, create a root user for you: username is root, password is 123456")
	hashedPassword, err := common.Password2Hash("123456")
	if err != nil {
		return err
	}
	rootUser := User{
		Username:    "root",
		Password:    hashedPassword,
		Role:        common.RoleRootUser,
		Status:      common.UserStatusEnabled,
		DisplayName: "超级管理员",
	}
	return DB.Create(&rootUser).Error
}

func CountTable(tableName string) (num int64) {
	DB.Table(tableName).Count(&num)
	return
}

func InitDB() (err error) {
	var db *gorm.DB
	if os.Getenv("SQL_DSN") != "" {
		// Use MySQL
		db, err = gorm.Open(mysql.Open(os.Getenv("SQL_DSN")), &gorm.Config{
			PrepareStmt: true, // precompile SQL
		})
	} else {
		// SQLite: PrepareStmt must stay off — it leaves statements open and breaks
		// AutoMigrate / Create with "cannot commit transaction - SQL statements in progress".
		db, err = gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
			PrepareStmt: false,
		})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr != nil {
				return dbErr
			}
			// SQLite does not handle concurrent writers well.
			sqlDB.SetMaxOpenConns(1)
			if pragmaErr := db.Exec("PRAGMA journal_mode=WAL;").Error; pragmaErr != nil {
				return pragmaErr
			}
			common.SysLog("SQL_DSN not set, using SQLite as database (WAL enabled)")
		}
	}
	if err == nil {
		DB = db
		err := db.AutoMigrate(&User{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Option{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Message{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Channel{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Webhook{})
		if err != nil {
			return err
		}
		err = createRootAccountIfNeed()
		return err
	} else {
		common.FatalLog(err)
	}
	return err
}

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}
