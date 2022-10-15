package main

import (
	. "common.bojiu.com/discover/kit/sd/etcdv3"
	"context"
	"fmt"
	"go.uber.org/zap"
	"yxxgame.bojiu.com/config"
	"yxxgame.bojiu.com/create"
	"yxxgame.bojiu.com/internal/gstream/server"
	"yxxgame.bojiu.com/pkg/log"
	"yxxgame.bojiu.com/pkg/mysql"
	"yxxgame.bojiu.com/pkg/redislib"
	"yxxgame.bojiu.com/pkg/viper"
)

func main() {

	// 初始化配置文件
	viper.InitVp()

	// 初始化日志文件
	log.ZapLog = log.InitLogger()

	// 初始化redis
	redislib.Sclient()

	// 初始化数据库 获取 mysql.M()  mysql.S()
	MasterDB := mysql.MasterInit()
	defer MasterDB.Close()
	Slave1DB := mysql.Slave1Init()
	defer Slave1DB.Close()

	log.ZapLog.Info("鱼虾蟹服务器开始........")
	/*******服务注册 start*******/
	scfg := config.NewServerCfg()
	instance := fmt.Sprintf("%s://%s:%d", scfg.GetProtocol(), scfg.GetIp(), scfg.GetPort())
	client, err := NewClient(context.Background(), scfg.GetEtcdServer(), scfg.GetOption())
	if err != nil {
		log.ZapLog.With(zap.Error(err), zap.Stack("trace")).Info("error")
	}

	// Build the registrar.
	registrar := NewRegistrar(client, Service{
		Key:   scfg.GetRegKey(),
		Value: instance,
	}, log.ZapLog)

	// Register our instance.
	registrar.Register()
	defer registrar.Deregister()
	v, _ := client.GetEntries(scfg.GetRegKey())
	log.ZapLog.With(zap.Any("regKey", v), zap.Stack("trace")).Info("main")
	/*******服务注册 end *******/
	go create.StartDesk()

	go server.Check()

	server.Run()

}
