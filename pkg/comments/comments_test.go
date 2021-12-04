package comments

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"os"
	"testing"
	"github.com/romanyx/polluter"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetCommentsByContent(t *testing.T) {
	boilerplate_testing.FlushPostgresTables()
}
