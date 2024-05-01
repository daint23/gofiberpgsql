package route

import (
	"sync"

	"github.com/daint23/gofiberpg/src/controller"
	"github.com/daint23/gofiberpg/src/repo"
	"github.com/daint23/gofiberpg/src/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ApiRoute(app *fiber.App, db *pgxpool.Pool, validate *validator.Validate, wg *sync.WaitGroup) {
	categoryRepository := repo.NewCategoryRepo(db)
	categoryService := service.NewCategoryService(categoryRepository, validate, wg)
	categoryController := controller.NewCategoryController(categoryService)

	api := app.Group("/api/v1")

	api.Post("/categories", categoryController.Insert)
	api.Get("/categories", categoryController.FindAll)
	api.Get("/categories/import", categoryController.ImportCsv)
	api.Get("/categories/:id", categoryController.FindById)
	api.Put("/categories/:id", categoryController.Update)
	api.Delete("/categories/:id", categoryController.Delete)
	api.Post("/categories/export", categoryController.ExportCsv)
	api.Post("/categories/exportgo", categoryController.ExportCsvGo)
}
