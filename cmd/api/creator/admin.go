package creator

import (
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/configs"
)

func InitAdminApi(adminEndpoint router.IRpcEndpoint, apiDef map[string]swagger.ApiDescription, cfg configs.Settings) error {
	return nil
}
