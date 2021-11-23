package boilerplate

import (
	"github.com/gocql/gocql"
	"time"
)

func GetScyllaCluster(config ScyllaConfiguration) *gocql.ClusterConfig {
	likeHandlerCluster := gocql.NewCluster(SplitHostsToSlice(config.Hosts)...)
	likeHandlerCluster.Keyspace = config.Keyspace

	likeHandlerCluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.UserName,
		Password: config.Password,
	}

	timeout := 10 * time.Second
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	likeHandlerCluster.Timeout = timeout
	likeHandlerCluster.MaxPreparedStmts = config.MaxPreparedStmts
	likeHandlerCluster.NumConns = config.NumConns
	likeHandlerCluster.MaxRoutingKeyInfo = config.MaxRoutingKeyInfo
	likeHandlerCluster.PageSize = config.PageSize

	return likeHandlerCluster
}
