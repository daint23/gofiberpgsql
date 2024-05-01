package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

func ConfigViper() *viper.Viper {
	viper := viper.New()

	viper.SetConfigFile(".env")
	errVi := viper.ReadInConfig()
	if errVi != nil {
		fmt.Println(errVi)
	}

	return viper
}
