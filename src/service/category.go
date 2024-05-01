package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"

	"github.com/daint23/gofiberpg/src/domain"
	"github.com/daint23/gofiberpg/src/helper"
	"github.com/daint23/gofiberpg/src/http/request"
	"github.com/daint23/gofiberpg/src/http/response"
	"github.com/daint23/gofiberpg/src/repo"
	"github.com/go-playground/validator/v10"
)

type CategoryService interface {
	Insert(ctx context.Context, req *request.CategoryCreateRequest) *response.CategoryResponse
	Update(ctx context.Context, req *request.CategoryUpdateRequest) *response.CategoryResponse
	Delete(ctx context.Context, categoryId int) error
	FindById(ctx context.Context, categoryId int) *response.CategoryResponse
	FindAll(ctx context.Context, params *request.CategoryQueryParams) []*response.CategoryResponse
	ExportCsv(ctx context.Context, head *multipart.FileHeader) error
	ImportCsv(ctx context.Context) error
	DispatchWorkers(jobs <-chan *domain.Category)
	OpenCsvFile(head *multipart.FileHeader) (*csv.Reader, error)
	ReadCsvFilePerLineThenSendToWorker(csvReader *csv.Reader, jobs chan<- *domain.Category)
}

type CategoryServiceImpl struct {
	CategoryRepo repo.CategoryRepo
	Validator    *validator.Validate
	Wg           *sync.WaitGroup
}

func NewCategoryService(categoryRepo repo.CategoryRepo, validator *validator.Validate, wg *sync.WaitGroup) CategoryService {
	return &CategoryServiceImpl{
		CategoryRepo: categoryRepo,
		Validator:    validator,
		Wg:           wg,
	}
}

// ImportCsvCsv implements CategoryService.
func (c *CategoryServiceImpl) ImportCsv(ctx context.Context) error {
	categories := c.CategoryRepo.ImportCsv(ctx)

	pathStorage := "./src/storage"
	errMkd := os.MkdirAll(pathStorage, 0755)
	if errMkd != nil {
		panic(helper.NewHTTPError(404, errMkd))
	}

	fileName := "output.csv"
	file, errCr := os.Create(filepath.Join(pathStorage, fileName))
	if errCr != nil {
		return errCr
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	errWr := writer.Write([]string{"name", "description"})
	if errWr != nil {
		return errWr
	}

	for _, category := range categories {
		errWri := writer.Write([]string{category.Name, category.Description})
		if errWri != nil {
			panic(helper.NewHTTPError(500, errors.New("woke")))
		}
	}

	return nil
}

func (c *CategoryServiceImpl) ExportCsv(ctx context.Context, head *multipart.FileHeader) error {
	file, errOpen := head.Open()
	if errOpen != nil {
		panic(helper.NewHTTPError(500, errOpen))
	}

	records, errRead := csv.NewReader(file).ReadAll()
	if errRead != nil {
		panic(helper.NewHTTPError(500, errRead))
	}

	// mengabaikan row pertama csv
	if len(records) > 0 {
		records = records[1:]
	}

	var valueStrings []string
	var valueArgs []interface{}

	totalCol := 2
	i := 0

	for _, record := range records {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d,$%d)", i*totalCol+1, i*totalCol+2))
		valueArgs = append(valueArgs, record[0], record[1])
		i++
	}

	errIn := c.CategoryRepo.ExportCsv(ctx, valueStrings, valueArgs)
	if errIn != nil {
		panic(helper.NewHTTPError(500, errIn))
	}

	return nil
}

// Delete implements CategoryService.
func (c *CategoryServiceImpl) Delete(ctx context.Context, categoryId int) error {
	findCategory := c.CategoryRepo.FindById(ctx, categoryId)

	err := c.CategoryRepo.Delete(ctx, findCategory.Id)
	if err != nil {
		return err
	}
	return nil
}

func (c *CategoryServiceImpl) FindAll(ctx context.Context, params *request.CategoryQueryParams) []*response.CategoryResponse {
	categories := c.CategoryRepo.FindAll(ctx, params)
	categoryResponses := []*response.CategoryResponse{}
	for _, category := range categories {
		categoryResponses = append(categoryResponses, &response.CategoryResponse{Id: category.Id, Name: category.Name, Description: category.Description})
	}
	return categoryResponses
}

// FindById implements CategoryService.
func (c *CategoryServiceImpl) FindById(ctx context.Context, categoryId int) *response.CategoryResponse {
	result := c.CategoryRepo.FindById(ctx, categoryId)
	return &response.CategoryResponse{Id: result.Id, Name: result.Name, Description: result.Description}
}

// Insert implements CategoryService.
func (c *CategoryServiceImpl) Insert(ctx context.Context, req *request.CategoryCreateRequest) *response.CategoryResponse {
	errVal := helper.ValidateStruct(req, c.Validator)
	if errVal != nil {
		panic(helper.NewHTTPInputValidationError(errVal))
	}

	category := &domain.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	result := c.CategoryRepo.Insert(ctx, category)
	return &response.CategoryResponse{
		Id:          result.Id,
		Name:        result.Name,
		Description: result.Description,
	}
}

// Update implements CategoryService.
func (c *CategoryServiceImpl) Update(ctx context.Context, req *request.CategoryUpdateRequest) *response.CategoryResponse {
	errVal := helper.ValidateStruct(req, c.Validator)
	if errVal != nil {
		panic(helper.NewHTTPInputValidationError(errVal))
	}

	findCategory := c.CategoryRepo.FindById(ctx, req.Id)

	category := &domain.Category{
		Id:          findCategory.Id,
		Name:        req.Name,
		Description: req.Description,
	}

	result := c.CategoryRepo.Update(ctx, category)
	return &response.CategoryResponse{
		Id:          result.Id,
		Name:        result.Name,
		Description: result.Description,
	}
}

func (service *CategoryServiceImpl) DispatchWorkers(jobs <-chan *domain.Category) {
	for workerIndex := 0; workerIndex <= 50; workerIndex++ {
		go func(workerIndex int, jobs <-chan *domain.Category) {
			counter := 0
			for job := range jobs {
				service.importData(workerIndex, counter, job)
				service.Wg.Done()
				counter++
			}
		}(workerIndex, jobs)
	}
}

func (service *CategoryServiceImpl) importData(workerIndex int, counter int, request *domain.Category) {
	for {
		var outerError error
		func(outerError *error) {
			defer func() {
				if err := recover(); err != nil {
					*outerError = fmt.Errorf("%v", err)
				}
			}()

			ctx := context.Background()

			errIn := service.CategoryRepo.ExportCsvGo(ctx, request)
			if errIn != nil {
				panic(helper.NewHTTPError(500, errors.New("insert")))
			}
		}(&outerError)
		if outerError == nil {
			break
		}
	}

	if counter%100 == 0 {
		log.Println("=> worker", workerIndex, "inserted", counter, "data")
	}
}

func (service *CategoryServiceImpl) OpenCsvFile(head *multipart.FileHeader) (*csv.Reader, error) {
	file, errOpen := head.Open()
	if errOpen != nil {
		panic(helper.NewHTTPError(500, errOpen))
	}

	reader := csv.NewReader(file)
	return reader, nil
}

func (service *CategoryServiceImpl) ReadCsvFilePerLineThenSendToWorker(csvReader *csv.Reader, jobs chan<- *domain.Category) {
	isHeader := true
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		if isHeader {
			isHeader = false
			continue
		}

		rowData := &domain.Category{
			Name:        row[0],
			Description: row[1],
		}

		service.Wg.Add(1)
		jobs <- rowData
	}

	close(jobs)
}
