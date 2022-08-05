package comments_counter

import (
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

func process(data eventData, executionData kafka_listener.ExecutionData) (*kafka.Message, error) {
	apm_helper.AddApmLabel(executionData.ApmTransaction, "content_id", data.ContentId)

	if err := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Exec("update creator_songs set comments = ? where id = ?",
		data.Count, data.ContentId).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &data.Messages, nil
}
