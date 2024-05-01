package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/daint23/gofiberpg/src/helper"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

func NewDB(viper *viper.Viper) *pgxpool.Pool {
	connection := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s", viper.GetString("POSTGRES_USER"), viper.GetString("POSTGRES_PASSWORD"), viper.GetString("POSTGRES_SERVICE"), viper.GetString("POSTGRES_DB"), viper.GetString("POSTGRES_SSL"))
	dbpool, err := pgxpool.New(context.Background(), connection)
	if err != nil {
		panic(helper.NewHTTPError(404, errors.New("hei")))
	}
	return dbpool
}
