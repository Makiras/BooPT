package main

import (
	"BooPT/config"
	"BooPT/database"
	r "BooPT/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {

	// read config file
	if config.Read("./config.yaml") != nil {
		logrus.Println("Error reading config file")
		os.Exit(1)
	}

	// get database connection
	if database.Connect() != nil {
		logrus.Println("Error connecting to database")
		os.Exit(1)
	}

	// init app
	app := fiber.New()
	app.Use(logger.New())

	// public routes
	setPublicRoutes(app)

	// jwt middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: config.JWTSALT,
		ContextKey: "jwt",
	}))

	// private routes
	setPrivateRoutes(app)

	err := app.Listen(":3000")
	if err != nil {
		logrus.Errorf("Error: %v ", err)
		os.Exit(1)
	}
}

func setPublicRoutes(app *fiber.App) {
	api := app.Group("/api")
	r.SetupAccountRouterPub(api)
}

func setPrivateRoutes(app *fiber.App) {
	api := app.Group("/api")
	r.SetupBookRouter(api)
	r.SetupLinkRouter(api)
	r.SetupTypeRouter(api)
	r.SetupTagRouter(api)
	r.SetupS3Router(api)
}
