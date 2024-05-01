package main

import (
	"sync"

	"github.com/daint23/gofiberpg/src/config"
	"github.com/daint23/gofiberpg/src/helper"
	"github.com/daint23/gofiberpg/src/route"
	"github.com/daint23/gofiberpg/src/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	viper := utils.ConfigViper()
	db := config.NewDB(viper)
	defer db.Close()
	validate := validator.New()

	config := fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Test Restapi",
		ErrorHandler:  helper.NewHTTPErrorHandler,
		BodyLimit:     5 * 1024 * 1024, /* 5MB */
	}

	configLog := logger.Config{
		Format:     "${pid} ${status} - ${time} ${latency} ${method} ${path}\n",
		TimeFormat: "02-01-2006",
		TimeZone:   "UTC",
		Output:     helper.LogDebug(),
	}

	app := fiber.New(config)

	app.Use(logger.New(configLog))
	app.Use(recover.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, PATCH, DELETE",
		AllowCredentials: true,
	}))

	wg := new(sync.WaitGroup)

	route.ApiRoute(app, db, validate, wg)

	wg.Wait()

	app.Listen(":8089")
}
