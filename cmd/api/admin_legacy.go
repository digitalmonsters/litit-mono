package api

import (
	"encoding/json"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/ads-manager/pkg/message"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/pkg/errors"
)

func InitLegacyAdminApi(adminLegacyEndpoint router.IRpcEndpoint, apiDef map[string]swagger.ApiDescription) error {
	if err := adminLegacyEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("UpsertMessageBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req message.UpsertMessageAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(req.Items) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("items are empty"), error_codes.GenericMappingError)
		}

		for _, i := range req.Items {
			if len(i.Countries) == 0 {
				return nil, error_codes.NewErrorWithCodeRef(errors.New("countries is required"), error_codes.GenericMappingError)
			}

			if len(i.Title) == 0 {
				return nil, error_codes.NewErrorWithCodeRef(errors.New("title is required"), error_codes.GenericMappingError)
			}

			if len(i.Description) == 0 {
				return nil, error_codes.NewErrorWithCodeRef(errors.New("description is required"), error_codes.GenericMappingError)
			}

			if i.AgeFrom == 0 {
				return nil, error_codes.NewErrorWithCodeRef(errors.New("age_from is required"), error_codes.GenericMappingError)
			}

			if i.AgeTo == 0 {
				return nil, error_codes.NewErrorWithCodeRef(errors.New("age_to is required"), error_codes.GenericMappingError)
			}

			if i.AgeFrom > i.AgeTo {
				return nil, error_codes.NewErrorWithCodeRef(errors.New("age_to is less than age_from"), error_codes.GenericMappingError)
			}

			if i.PointsFrom > 0 || i.PointsTo > 0 {
				if i.PointsFrom == 0 || i.PointsTo == 0 {
					return nil, error_codes.NewErrorWithCodeRef(errors.New("if you need points condition, points_from and points_to is required"), error_codes.GenericMappingError)
				}
			}
		}

		resp, err := message.UpsertMessageBulkAdmin(req, database.GetDbWithContext(database.DbTypeMaster, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})); err != nil {
		return err
	}

	if err := adminLegacyEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("DeleteMessagesBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req message.DeleteMessagesBulkAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(req.Ids) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("ids are empty"), error_codes.GenericMappingError)
		}

		if err := message.DeleteMessagesBulkAdmin(req, database.GetDbWithContext(database.DbTypeMaster, executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	})); err != nil {
		return err
	}

	if err := adminLegacyEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("MessagesListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req message.MessagesListAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := message.MessagesListAdmin(req, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})); err != nil {
		return err
	}

	apiDef["UpsertMessageBulkAdmin"] = swagger.ApiDescription{
		Request:  message.UpsertMessageAdminRequest{},
		Response: []database.Message{},
		Tags:     []string{"message", "upsert"},
	}

	apiDef["DeleteMessagesBulkAdmin"] = swagger.ApiDescription{
		Request:  message.DeleteMessagesBulkAdminRequest{},
		Response: nil,
		Tags:     []string{"message", "delete"},
	}

	apiDef["MessagesListAdmin"] = swagger.ApiDescription{
		Request:  message.MessagesListAdminRequest{},
		Response: message.MessagesListAdminResponse{},
		Tags:     []string{"message", "list"},
	}

	return nil
}
