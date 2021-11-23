package boilerplate

import "github.com/bwmarrin/snowflake"

var idGen *snowflake.Node

func GetGenerator() *snowflake.Node {
	if idGen != nil {
		return idGen
	}

	idGen, _ = snowflake.NewNode(1)

	return idGen
}
