package configs

import (
	"bytes"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load(configMap any, defaultConfig []byte) error {
	viper.SetConfigType("yaml")

	err := viper.ReadConfig(bytes.NewBuffer(defaultConfig))
	if err != nil {
		return err
	}

	// Load envs in file .env into env of go
	godotenv.Load(".env")

	// Load env to override default value
	viper.AutomaticEnv()

	// fmt.Println("DB Host:", viper.GetString("database_postgres.host"))
	// fmt.Println("DB Port:", viper.GetInt("database_postgres.port"))
	// Support viper to know get correct value from env value
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	// viper.GetString("database_postgres.host") to viper.GetString("database_postgres__host")

	err = viper.Unmarshal(&configMap)
	if err != nil {
		return err
	}

	return nil
}
