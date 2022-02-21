package database

import (
	. "BooPT/config"
	m "BooPT/model"
	"fmt"
	gorm_logrus "github.com/onrik/gorm-logrus"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func migrate() error {
	flag := false
	flag = DB.AutoMigrate(m.User{}) != nil || flag
	flag = DB.AutoMigrate(m.Book{}) != nil || flag
	flag = DB.AutoMigrate(m.Type{}) != nil || flag
	flag = DB.AutoMigrate(m.Permission{}) != nil || flag
	flag = DB.AutoMigrate(m.Tag{}) != nil || flag
	flag = DB.AutoMigrate(m.RequestRecord{}) != nil || flag
	flag = DB.AutoMigrate(m.DownloadRecord{}) != nil || flag
	flag = DB.AutoMigrate(m.PublishRecord{}) != nil || flag
	flag = DB.AutoMigrate(m.UploadRecord{}) != nil || flag
	flag = DB.AutoMigrate(m.FavoriteRecord{}) != nil || flag
	flag = DB.AutoMigrate(m.DownloadLink{}) != nil || flag
	if flag {
		logrus.Errorf("migrate error")
		return gorm.ErrInvalidDB
	}
	return nil
}

func Connect() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		CONFIG.Database.Host,
		CONFIG.Database.Port,
		CONFIG.Database.User,
		CONFIG.Database.DbName,
		CONFIG.Database.Password)
	logrus.Infof("Connecting to database with dsn: %v ", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gorm_logrus.New(),
	})
	if err != nil {
		logrus.Errorf("Error connecting to database: %v ", err)
		return err
	}
	DB = db
	//err = migrate()
	return err
}
