package server

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"yxxgame.bojiu.com/internal/gstream/proto"
	"yxxgame.bojiu.com/pkg/log"
)

func GoClient() proto.StorageClient {
	//log.ZapLog = log.InitLogger()
	//scfg := config.NewServerCfg()
	//log.ZapLog.Info(fmt.Sprintf("%s:%d", scfg.GetIp(), scfg.GetPort()))

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "127.0.0.1", 19001), grpc.WithInsecure())
	if err != nil {
		log.ZapLog.With(zap.Error(err)).Error("grpc dial error")
	}

	return proto.NewStorageClient(conn)
}
