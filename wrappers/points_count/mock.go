package points_count

import "go.elastic.co/apm"

//goland:noinspection ALL
type PointsCountWrapperMock struct {
	GetPointsCountFn func(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetPointsCountResponseChan
}

func (m *PointsCountWrapperMock) GetPointsCount(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetPointsCountResponseChan {
	return m.GetPointsCountFn(contentIds, apmTransaction, forceLog)
}

func GetMock() IPointsCountWrapper { // for compiler errors
	return &PointsCountWrapperMock{}
}
