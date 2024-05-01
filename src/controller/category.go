package controller

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/daint23/gofiberpg/src/domain"
	"github.com/daint23/gofiberpg/src/helper"
	"github.com/daint23/gofiberpg/src/http/request"
	"github.com/daint23/gofiberpg/src/service"
	"github.com/gofiber/fiber/v2"
)

type CategoryController interface {
	Insert(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
	FindById(ctx *fiber.Ctx) error
	FindAll(ctx *fiber.Ctx) error
	ExportCsv(ctx *fiber.Ctx) error
	ImportCsv(ctx *fiber.Ctx) error
	ExportCsvGo(ctx *fiber.Ctx) error
}

type CategoryControllerImpl struct {
	CategoryService service.CategoryService
}

func NewCategoryController(categoryService service.CategoryService) CategoryController {
	return &CategoryControllerImpl{
		CategoryService: categoryService,
	}
}

// ImportCsv implements CategoryController.
func (c *CategoryControllerImpl) ImportCsv(ctx *fiber.Ctx) error {
	ctx.Set("Content-Type", "text/csv")
	ctx.Set("Content-Disposition", "attachment; filename=output.csv")
	err := c.CategoryService.ImportCsv(ctx.Context())
	if err != nil {
		panic(helper.NewHTTPError(500, err))
	}

	pathStorage := "./src/storage"
	defer os.Remove(filepath.Join(pathStorage, "output.csv"))

	return ctx.Download("./src/storage/output.csv")
}

// ExportCsv implements CategoryController.
func (c *CategoryControllerImpl) ExportCsv(ctx *fiber.Ctx) error {
	head, err := ctx.FormFile("file")
	if err != nil {
		panic(helper.NewHTTPError(500, err))
	}

	errBatch := c.CategoryService.ExportCsv(ctx.Context(), head)
	if errBatch != nil {
		panic(helper.NewHTTPError(500, errBatch))
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success export csv"})
}

// Delete implements CategoryController.
func (c *CategoryControllerImpl) Delete(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		panic(helper.NewHTTPError(404, errors.New("id not found")))
	}

	errDel := c.CategoryService.Delete(ctx.Context(), id)
	if errDel != nil {
		panic(helper.NewHTTPError(500, errDel))
	}

	return ctx.Status(300).JSON(fiber.Map{"message": "success"})
}

// FindAll implements CategoryController.
func (c *CategoryControllerImpl) FindAll(ctx *fiber.Ctx) error {
	params := &request.CategoryQueryParams{}
	err := ctx.QueryParser(params)
	if err != nil {
		panic(helper.NewHTTPError(500, errors.New("Ooppss")))
	}
	result := c.CategoryService.FindAll(ctx.Context(), params)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
}

// FindById implements CategoryController.
func (c *CategoryControllerImpl) FindById(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		panic(helper.NewHTTPError(404, errors.New("id not found")))
	}

	result := c.CategoryService.FindById(ctx.Context(), id)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
}

// Insert implements CategoryController.
func (c *CategoryControllerImpl) Insert(ctx *fiber.Ctx) error {
	req := &request.CategoryCreateRequest{}
	err := ctx.BodyParser(req)
	if err != nil {
		panic(helper.NewHTTPError(fiber.StatusInternalServerError, errors.New("body is required")))
	}
	result := c.CategoryService.Insert(ctx.Context(), req)
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"data": result})
}

// Update implements CategoryController.
func (c *CategoryControllerImpl) Update(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		panic(helper.NewHTTPError(404, errors.New("id not found")))
	}

	req := &request.CategoryUpdateRequest{}
	errPar := ctx.BodyParser(req)
	if errPar != nil {
		panic(helper.NewHTTPError(fiber.StatusInternalServerError, errors.New("body is required")))
	}

	req.Id = id

	result := c.CategoryService.Update(ctx.Context(), req)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
}

func (controller *CategoryControllerImpl) ExportCsvGo(ctx *fiber.Ctx) error {
	head, err := ctx.FormFile("file")
	if err != nil {
		panic(helper.NewHTTPError(500, err))
	}
	reader, errOpen := controller.CategoryService.OpenCsvFile(head)
	if errOpen != nil {
		panic(helper.NewHTTPError(500, errOpen))
	}

	jobs := make(chan *domain.Category)
	go controller.CategoryService.DispatchWorkers(jobs)
	controller.CategoryService.ReadCsvFilePerLineThenSendToWorker(reader, jobs)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "oke"})
}
