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
	"testing"
	"time"
)

func GetScyllaTestCluster(config *boilerplate.ScyllaConfiguration) (*gocql.ClusterConfig, *gocql.Session, error) {
	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci {
		config.Keyspace = fmt.Sprintf("a%v", boilerplate.GetGenerator().Generate().String())
	}
	oldKeyspace := config.Keyspace

	if oldKeyspace != "system" {
		if err := ensureKeyspaceExists(*config, oldKeyspace); err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	return getScyllaClusterInternal(*config)
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

	for _, tbl := range tables {

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

func ensureKeyspaceExists(config boilerplate.ScyllaConfiguration, targetKeyspace string) error {
	config.Keyspace = "system"

	_, ses, err := getScyllaClusterInternal(config)

	defer ses.Close()

	if err != nil {
		return errors.WithStack(err)
	}

	iter := ses.Query(fmt.Sprintf("SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name='%v'", targetKeyspace)).Iter()

	var keyspaceName string

	for iter.Scan(&keyspaceName) {
		break
	}

	if err := iter.Close(); err != nil {
		return err
	}

	if keyspaceName == targetKeyspace {
		return nil
	}

	if err = ses.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %v WITH replication = {
			'class' : 'SimpleStrategy',
			'replication_factor' : 1
		}`, targetKeyspace)).Exec(); err != nil {
		return err
	}

	return err
}

func getScyllaClusterInternal(config boilerplate.ScyllaConfiguration) (*gocql.ClusterConfig, *gocql.Session, error) {
	newCluster := gocql.NewCluster(boilerplate.SplitHostsToSlice(config.Hosts)...)

	timeout := 20 * time.Second

	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	newCluster.Timeout = timeout
	newCluster.ConnectTimeout = timeout
	newCluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.UserName,
		Password: config.Password,
	}

	newCluster.MaxPreparedStmts = config.MaxPreparedStmts
	newCluster.NumConns = config.NumConns
	newCluster.MaxRoutingKeyInfo = config.MaxRoutingKeyInfo
	newCluster.PageSize = config.PageSize
	newCluster.Keyspace = config.Keyspace

	newSession, err := newCluster.CreateSession()

	return newCluster, newSession, err
}
