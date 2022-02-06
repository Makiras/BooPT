package config

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type DatabaseConf struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
}

type LDAPConf struct {
	Addr     string
	User     string
	Password string
	BaseDN   string
}

type Config struct {
	Database DatabaseConf
	LDAP     LDAPConf
	LogLevel int // 0: error, 1: warn, 2: info, 3: debug
	JwtSalt  string
}

var CONFIG = Config{
	Database: DatabaseConf{
		Host:     "localhost",
		Port:     5432,
		User:     "BooPT",
		Password: "BooPTPassword",
		DbName:   "BooPT",
	},
	LDAP: LDAPConf{
		Addr:     "localhost:389",
		User:     "cn=admin,dc=example,dc=com",
		Password: "admin",
		BaseDN:   "dc=example,dc=com",
	},
	LogLevel: 3,
	JwtSalt:  "BooPTJWTSalt",
}

var JWTSALT = []byte(CONFIG.JwtSalt)

func Read(filename string) error {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Errorf("yamlFile.Get err %#v ", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &CONFIG)
	if err != nil {
		logrus.Errorf("Unmarshal: %#v ", err)
		return err
	}
	return nil
}
