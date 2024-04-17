package boilerplate

import (
	"fmt"
)

func GetDbConnectionString(config DbConfig) (string, error) {
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", config.Host, config.User, config.Password, config.Db, config.Port), nil
}
