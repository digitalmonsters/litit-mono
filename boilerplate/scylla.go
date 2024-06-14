package boilerplate

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm/module/apmgocql/v2"
)

func GetScyllaCluster(config ScyllaConfiguration) *gocql.ClusterConfig {
	oldKeyspace := config.Keyspace

	if oldKeyspace != "system" {
		if err := EnsureKeyspaceExists(config, oldKeyspace); err != nil {
			log.Panic().Err(errors.WithStack(err)).Str("host", config.Hosts).Str("keyspace", config.Keyspace).
				Msg("[Scylla Cluster] : create 'system' keyspace failed")
		}
	}

	cluster := GetScyllaClusterInternal(config)

	observer := apmgocql.NewObserver()
	cluster.QueryObserver = observer
	cluster.BatchObserver = observer

	return cluster
}

func EnsureKeyspaceExists(config ScyllaConfiguration, targetKeyspace string) error {
	config.Keyspace = "system"

	cluster := GetScyllaClusterInternal(config)

	ses, err := cluster.CreateSession()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "create session")
	}

	defer ses.Close()

	iter := ses.Query(fmt.Sprintf("SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name='%v'", targetKeyspace)).Iter()

	var keyspaceName string

	for iter.Scan(&keyspaceName) {
		break
	}

	if err := iter.Close(); err != nil {
		return errors.Wrap(errors.WithStack(err), "close iterator")
	}

	if keyspaceName == targetKeyspace {
		return nil
	}

	if err = ses.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %v WITH replication = {
			'class' : 'SimpleStrategy',
			'replication_factor' : 1
		}`, targetKeyspace)).Exec(); err != nil {
		return errors.Wrap(errors.WithStack(err), "query exec")
	}

	return err
}

func GetScyllaClusterInternal(config ScyllaConfiguration) *gocql.ClusterConfig {
	newCluster := gocql.NewCluster(SplitHostsToSlice(config.Hosts)...)

	timeout := 20 * time.Second

	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	newCluster.Timeout = timeout
	//newCluster.ConnectTimeout = timeout
	newCluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.UserName,
		Password: config.Password,
	}

	newCluster.MaxPreparedStmts = config.MaxPreparedStmts
	newCluster.NumConns = config.NumConns
	newCluster.MaxRoutingKeyInfo = config.MaxRoutingKeyInfo
	newCluster.PageSize = config.PageSize
	newCluster.Keyspace = config.Keyspace

	return newCluster
}
