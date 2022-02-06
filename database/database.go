package database

import (
	. "BooPT/config"
	"fmt"
	gorm_logrus "github.com/onrik/gorm-logrus"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

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
	return nil
}
