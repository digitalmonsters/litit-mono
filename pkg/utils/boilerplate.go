package utils

import (
	"bytes"
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/romanyx/polluter"
	"gorm.io/gorm"
	"io/ioutil"
	"time"
)

func PollutePostgresDatabase(gormDb *gorm.DB, filePaths ...string) error {
	var found []string

	for _, fileToFind := range filePaths {
		if path, err := boilerplate.RecursiveFindFile(fileToFind, "./", 30); err != nil {
			return errors.WithStack(err)
		} else {
			found = append(found, path)
		}
	}

	db, err := gormDb.DB()
	if err != nil {
		return err
	}

	seeder := polluter.
		New(polluter.JSONParser, polluter.PostgresEngine(db))

	for _, f := range found {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		if err = seeder.Pollute(bytes.NewReader(data)); err != nil {
			return err
		}
	}

	return nil
}

var localCache = cache.New(1*time.Minute, 1*time.Minute)

func GetUser(userId int64, ctx context.Context) (*scylla.User, error) {
	if userId == 0 {
		return nil, errors.WithStack(errors.New("user not found"))
	}

	cacheKey := fmt.Sprintf("user_%v", userId)

	if v, ok := localCache.Get(cacheKey); ok {
		return v.(*scylla.User), nil
	}

	session := database.GetScyllaSession()

	var user scylla.User

	userIter := session.Query("select user_id, username, firstname, lastname, name_privacy_status, language "+
		"from user where cluster_key = ? and user_id = ?", scylla.GetUserClusterKey(userId), userId).WithContext(ctx).Iter()
	userIter.Scan(&user.UserId, &user.Username, &user.Firstname, &user.Lastname, &user.NamePrivacyStatus, &user.Language)

	if err := userIter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	localCache.SetDefault(cacheKey, user)

	return &user, nil
}

func GetUserRenderingVariablesWithLanguage(userId int64, ctx context.Context) (database.RenderingVariables, translation.Language, error) {
	if user, err := GetUser(userId, ctx); err != nil {
		return database.RenderingVariables{}, translation.DefaultUserLanguage, errors.WithStack(err)
	} else {
		userRecord := user_go.UserRecord{
			UserId:            userId,
			Username:          user.Username,
			Firstname:         user.Firstname,
			Lastname:          user.Lastname,
			NamePrivacyStatus: user.NamePrivacyStatus,
		}

		firstname, lastname := userRecord.GetFirstAndLastNameWithPrivacy()

		renderingVariables := database.RenderingVariables{
			"firstname": firstname,
			"lastname":  lastname,
		}

		return renderingVariables, user.Language, nil
	}
}
