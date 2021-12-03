package publicapi

import "gorm.io/gorm"

func GetCommendById(commentId int64, db *gorm.DB) (interface{}, error) {

}

func DeleteCommentById(commentId int64, db *gorm.DB) (interface{}, error) {

}

func UpdateCommentById(commentId int64, comment string, db *gorm.DB) (interface{}, error) {

}

func GetRepliesByCommentId(commentId int64, db *gorm.DB) (interface{}, error) {

}

func VoteComment(commentId int64, voteUp bool, db *gorm.DB) (interface{}, error) {

}

func ReportComment(commentId int64, details string, db *gorm.DB) (interface{}, error) {

}

func GetCommentByTypeWithResourceId(commentType string, resourceId int64, db *gorm.DB) (interface{}, error) {

}

func SendComment(commentType string, resourceId int64, db *gorm.DB) (interface{}, error) {

}
