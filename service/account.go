package service

import (
	"BooPT/config"
	db "BooPT/database"
	m "BooPT/model"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgconn"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
	"regexp"
	"strconv"
	"time"
)

var emailReg = regexp.MustCompile("\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*")

func ldapAuth(stuId, stuPassword string) (bool, error) {

	//return true, nil // just for debug

	ldapConfig := config.CONFIG.LDAP
	ldapConn, err := ldap.Dial("tcp", ldapConfig.Addr)
	if err != nil {
		return false, err
	}
	defer ldapConn.Close()

	err = ldapConn.Bind(ldapConfig.User, ldapConfig.Password)
	if err != nil {
		return false, err
	}

	searchRequest := ldap.NewSearchRequest(
		ldapConfig.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(uid=%s)", stuId),
		[]string{"dn", "cn", "mail"},
		nil,
	)

	sr, err := ldapConn.Search(searchRequest)
	if err != nil {
		return false, err
	}

	if len(sr.Entries) != 1 {
		return false, nil
	}

	userDN := sr.Entries[0].DN
	err = ldapConn.Bind(userDN, stuPassword)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func jwtSign(user m.User) (string, error) {
	if user.State == -1 {
		return "You have been banned!", nil
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.ID,
		"stu_id": user.StuId,
		"name":   user.Name,
		"email":  user.Email,
		"state":  user.State,
		"iat":    user.CreatedAt.Unix(),
		"exp":    user.CreatedAt.Add(time.Hour * 72).Unix(),
	})
	return rawToken.SignedString(config.JWTSALT)
}

func Login(c *fiber.Ctx) error {
	user := c.FormValue("user")
	pass := c.FormValue("password")
	if user == "" || pass == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing username or password",
		})
	}
	var userData m.User
	if emailReg.MatchString(user) { // if user use email to login
		if err := db.DB.Where("email = ?", user).First(&userData).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}
		// use base64ed student id to login
		bStuId := make([]byte, 8)
		binary.LittleEndian.PutUint64(bStuId, uint64(userData.StuId))
		salt := base64.StdEncoding.EncodeToString(bStuId)
		if userData.PasswordHash == base64.StdEncoding.EncodeToString(argon2.IDKey([]byte(pass), []byte(salt), 1, 64*1024, 1, 32)) {
			// sign JWT
			token, err := jwtSign(userData)
			if err != nil {
				logrus.Errorf("Failed to generate token: %#v", err)
				return c.Status(500).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
			return c.JSON(fiber.Map{
				"token": token,
			})
		}
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	} else { // else if user use student id to login

		// is student id valid?
		userId, err := strconv.ParseInt(user, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid student ID or password",
			})
		}
		// is password correct?
		if ldapRes, err := ldapAuth(user, pass); err != nil {
			logrus.Errorf("Failed to ldap auth: %#v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		} else if ldapRes {
			// is user exist in database?
			if err := db.DB.Where(m.User{StuId: userId}).First(&userData).Error; err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": "Invalid Student ID or password",
				})
			}
			// sign JWT
			token, err := jwtSign(userData)
			if err != nil {
				logrus.Errorf("Failed to generate token: %#v", err)
				return c.Status(500).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
			return c.JSON(fiber.Map{
				"token": token,
			})
		}
		// if password is incorrect
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Student ID or password",
		})
	}
}

func Register(c *fiber.Ctx) error {
	email := c.FormValue("email")
	userName := c.FormValue("name")
	password := c.FormValue("password")
	studentID := c.FormValue("stu_id")
	ldapPass := c.FormValue("ldap_password")

	// ldap validation
	userId, err := strconv.ParseInt(studentID, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Student ID",
		})
	}
	if ldapRes, err := ldapAuth(studentID, ldapPass); err != nil {
		logrus.Errorf("Failed to ldap auth: %#v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	} else if ldapRes {
		bStuId := make([]byte, 8)
		binary.LittleEndian.PutUint64(bStuId, uint64(userId))
		salt := base64.StdEncoding.EncodeToString(bStuId)
		userData := m.User{
			Email:        email,
			Name:         userName,
			StuId:        userId,
			PasswordHash: base64.StdEncoding.EncodeToString(argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 1, 32)),
		}
		result := db.DB.Create(&userData)
		if result.Error != nil {
			logrus.Errorf("Failed to create user %d due to %#v", userId, result.Error)
			var pgErr *pgconn.PgError
			if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
				// duplicate key value violates unique constraint
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Your student id has been used.",
				})
			}
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
		if token, err := jwtSign(userData); err != nil {
			logrus.Errorf("Failed to sign jwt due to %#v", result.Error)
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error while creat JWT, but your account has been created.",
			})
		} else {
			return c.JSON(fiber.Map{
				"token": token,
			})
		}
	}

	return c.Status(400).JSON(fiber.Map{
		"error": "Invalid Student ID or password",
	})
}
