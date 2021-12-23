package boilerplate

import (
	"github.com/gocql/gocql"
	"time"
)

func GetScyllaCluster(config ScyllaConfiguration) *gocql.ClusterConfig {
	//observer := apmgocql.NewObserver()
	scyllaCluster := gocql.NewCluster(SplitHostsToSlice(config.Hosts)...)
	scyllaCluster.Keyspace = config.Keyspace

	scyllaCluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.UserName,
		Password: config.Password,
	}

	timeout := 10 * time.Second
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	scyllaCluster.Timeout = timeout
	scyllaCluster.MaxPreparedStmts = config.MaxPreparedStmts
	scyllaCluster.NumConns = config.NumConns
	scyllaCluster.MaxRoutingKeyInfo = config.MaxRoutingKeyInfo
	scyllaCluster.PageSize = config.PageSize

	//scyllaCluster.QueryObserver = observer
	//scyllaCluster.BatchObserver = observer

	return scyllaCluster
}
