package boilerplate_testing

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/scylla_migrator"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/scylladb/gocqlx/v2"
	"github.com/thoas/go-funk"
	"io/fs"
	"os"
	"testing"
)

func GetScyllaTestCluster(config *boilerplate.ScyllaConfiguration) (*gocql.ClusterConfig, *gocql.Session, error) {
	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci {
		config.Keyspace = GetScyllaCiKeyspaceName()
	}
	oldKeyspace := config.Keyspace

	if oldKeyspace != "system" {
		if err := boilerplate.EnsureKeyspaceExists(*config, oldKeyspace); err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	return boilerplate.GetScyllaClusterInternal(*config)
}

func GetScyllaCiKeyspaceName() string {
	return fmt.Sprintf("ci_%v", int64(os.Getpid()))
}

func RunScyllaMigrations(session *gocql.Session, migrations fs.FS) error {
	wrappedSession, err := gocqlx.WrapSession(session, nil)

	if err != nil {
		return errors.WithStack(err)
	}

	return scylla_migrator.FromFS(context.TODO(), wrappedSession, migrations)
}

func FlushScyllaAllTables(t *testing.T, session *gocql.Session, keyspace string, except []string) error {
	except = append(except, "gocqlx_migrate")

	return innerScyllaFlushTables(t, session, keyspace, except, nil)
}

func FlushScyllaTables(t *testing.T, session *gocql.Session, keyspace string, exact []string) error {
	except := []string{"gocqlx_migrate"}

	return innerScyllaFlushTables(t, session, keyspace, except, exact)
}

func innerScyllaFlushTables(t *testing.T, session *gocql.Session, keyspace string, except, exact []string) error {
	metaData, err := session.KeyspaceMetadata(keyspace)

	if err != nil {
		return err
	}

	var tables = metaData.Tables

	var views []string

	for _, v := range metaData.MaterializedViews {
		views = append(views, v.Name)
	}

	for _, tbl := range tables {
		if funk.ContainsString(views, tbl.Name) {
			continue
		}

		tblFullName := fmt.Sprintf("%s.%s;", keyspace, tbl.Name)

		if funk.Contains(except, tblFullName) || funk.Contains(except, tbl.Name) {
			continue
		}

		if len(exact) == 0 {
			q := fmt.Sprintf("truncate %s;", tblFullName)

			if err := session.Query(q).Exec(); err != nil {
				return err
			}
		} else {
			for _, ex := range exact {
				if ex == tbl.Name || ex == tblFullName {
					q := fmt.Sprintf("truncate %s;", tblFullName)

					if err := session.Query(q).Exec(); err != nil {
						return err
					}
				}
			}
		}

	}
	return nil
}
