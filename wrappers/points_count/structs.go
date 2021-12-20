package points_count

import (
	"github.com/digitalmonsters/go-common/rpc"
)

type GetPointsCountResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]PointsCountRecord `json:"items"`
}

//goland:noinspection GoNameStartsWithPackageName
type PointsCountRecord struct {
	Points float64 `json:"points"`
}

type GetPointsCountRequest struct {
	ContentIds []int64 `json:"ids"`
}
