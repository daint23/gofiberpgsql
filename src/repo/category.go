package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/daint23/gofiberpg/src/domain"
	"github.com/daint23/gofiberpg/src/helper"
	"github.com/daint23/gofiberpg/src/http/request"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepo interface {
	Insert(ctx context.Context, category *domain.Category) *domain.Category
	Update(ctx context.Context, category *domain.Category) *domain.Category
	Delete(ctx context.Context, categoryId int) error
	FindById(ctx context.Context, categoryId int) *domain.Category
	FindAll(ctx context.Context, params *request.CategoryQueryParams) []*domain.Category
	ExportCsv(ctx context.Context, valueStrings []string, valueArgs []interface{}) error
	ImportCsv(ctx context.Context) []*domain.Category
	ExportCsvGo(ctx context.Context, request *domain.Category) error
}

type CategoryRepoImpl struct {
	DB *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) CategoryRepo {
	return &CategoryRepoImpl{
		DB: db,
	}
}

// ImportCsv implements CategoryRepo.
func (c *CategoryRepoImpl) ImportCsv(ctx context.Context) []*domain.Category {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)

	SQL := "select name,description from category"
	rows, err := tx.Query(ctx, SQL)
	if err != nil {
		panic(helper.NewHTTPError(500, errors.New("Ooppss")))
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		errScan := rows.Scan(&category.Name, &category.Description)
		if errScan != nil {
			panic(helper.NewHTTPError(500, errScan))
		}
		categories = append(categories, category)
	}
	return categories
}

func (c *CategoryRepoImpl) ExportCsv(ctx context.Context, valueStrings []string, valueArgs []interface{}) error {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)

	SQL := "insert into category (name,description) values "
	SQL += strings.Join(valueStrings, ",")
	_, errExec := tx.Exec(ctx, SQL, valueArgs...)
	if errExec != nil {
		return errExec
	}

	return nil
}

// Delete implements CategoryRepo.
func (c *CategoryRepoImpl) Delete(ctx context.Context, categoryId int) error {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)

	SQL := "delete from category where id = $1 returning id"
	_, errEx := tx.Exec(ctx, SQL, categoryId)
	if errEx != nil {
		return errEx
	}

	return nil
}

// FindAll implements CategoryRepo.
func (c *CategoryRepoImpl) FindAll(ctx context.Context, params *request.CategoryQueryParams) []*domain.Category {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)
	SQL := "select id,name,description from category where id > $1 order by id asc limit $2"
	rows, err := tx.Query(ctx, SQL, params.Id, params.Limit)
	if err != nil {
		panic(helper.NewHTTPError(500, errors.New("Ooppss")))
	}

	defer rows.Close()

	var categories []*domain.Category

	for rows.Next() {
		category := &domain.Category{}
		errScan := rows.Scan(&category.Id, &category.Name, &category.Description)
		if errScan != nil {
			panic(helper.NewHTTPError(500, errScan))
		}
		categories = append(categories, category)
	}
	return categories
}

// FindById implements CategoryRepo.
func (c *CategoryRepoImpl) FindById(ctx context.Context, categoryId int) *domain.Category {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)

	SQL := "select id, name, description from category where id = $1"
	category := &domain.Category{}
	errQuery := tx.QueryRow(ctx, SQL, categoryId).Scan(&category.Id, &category.Name, &category.Description)
	if errQuery != nil {
		panic(helper.NewHTTPError(404, errors.New("category not found")))
	}

	return category
}

// Insert implements CategoryRepo.
func (c *CategoryRepoImpl) Insert(ctx context.Context, category *domain.Category) *domain.Category {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)

	SQL := "insert into category(name, description) values($1, $2) returning id"
	err := tx.QueryRow(ctx, SQL, category.Name, category.Description).Scan(&category.Id)
	if err != nil {
		panic(helper.NewHTTPError(500, err))
	}

	return category
}

// Update implements CategoryRepo.
func (c *CategoryRepoImpl) Update(ctx context.Context, category *domain.Category) *domain.Category {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)

	SQL := "update category set name = $1, description = $2 where id = $3 returning id,name,description"
	err := tx.QueryRow(ctx, SQL, category.Name, category.Description, category.Id).Scan(&category.Id, &category.Name, &category.Description)
	if err != nil {
		panic(helper.NewHTTPError(500, errors.New("Ooppss")))
	}
	return category
}

// ExportCsvGo implements CategoryRepo.
func (c *CategoryRepoImpl) ExportCsvGo(ctx context.Context, request *domain.Category) error {
	tx, errBegin := c.DB.Begin(ctx)
	if errBegin != nil {
		panic(helper.NewHTTPError(500, errBegin))
	}
	defer helper.CommitOrRollback(tx)

	data := []interface{}{
		request.Name,
		request.Description,
	}

	SQL := fmt.Sprintf("insert into category (name,description) values (%s)",
		generateDollarsMark(data),
	)

	_, errExec := tx.Exec(ctx, SQL, data...)
	if errExec != nil {
		panic(helper.NewHTTPError(500, errors.New("exec")))
	}

	return nil
}

func generateDollarsMark(data []interface{}) string {
	s := make([]string, 0)

	for i := 1; i <= len(data); i++ {
		s = append(s, fmt.Sprintf("$%d", i))
	}

	return strings.Join(s, ",")
}
