package server

import (
	"common.bojiu.com/def"
	Utils "common.bojiu.com/utils"
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net"
	"strings"
	"time"
	"yxxgame.bojiu.com/config"
	c "yxxgame.bojiu.com/create"
	"yxxgame.bojiu.com/enum"
	"yxxgame.bojiu.com/internal/gstream/pb"
	cproto "yxxgame.bojiu.com/internal/gstream/proto"
	"yxxgame.bojiu.com/model"
	"yxxgame.bojiu.com/pkg/log"
	protoStructure "yxxgame.bojiu.com/proto"
)

// var center  = storage.StorageServerImpl
var stream *streamServer

type streamServer struct {
	GrpcRecvClientData chan *pb.StreamRequestData
	GrpcSendClientData chan *pb.StreamResponseData
}

func NewStreamServer() *streamServer {
	stream = &streamServer{
		make(chan *pb.StreamRequestData, 100),
		make(chan *pb.StreamResponseData, 100),
	}
	//
	return stream
}

//func init() {
//	GrpcRecvClientData = make(chan *pb.StreamRequestData, 100)
//	GrpcSendClientData = make(chan *pb.StreamResponseData, 100)
//}

// PPStream log.ZapLog.With(zap.Any("err", err)).Error("收到网关数据错误")
func (gs *streamServer) PPStream(stream pb.ForwardMsg_PPStreamServer) error {
	stop := make(chan struct{})
	defer func() {
		if e := recover(); e != nil {
			log.ZapLog.Info("PPStream recover", zap.Any("err", e.(error)))
		}
		close(stop)
	}()
	go gs.response(stream, stop)
	go gs.dispatch(stop)
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.ZapLog.Info("PPStream recv io EOF", zap.Any("err", err))
			return nil
		}
		if err != nil {
			log.ZapLog.Info("PPStream recv error", zap.Any("err", err))
			return err
		}
		info := fmt.Sprintf("收到网关数据:协议号=%+v,加密字符=%+v,随机字符=%+v,protobuf=%+v", msg.GetMsg(), Utils.ToHexString(msg.GetSecret()), msg.GetSerialNum(), msg.GetData())
		log.ZapLog.Info(info)
		gs.GrpcRecvClientData <- msg
	}
}

func (gs *streamServer) response(stream pb.ForwardMsg_PPStreamServer, stop chan struct{}) {
	defer func() {
		if e := recover(); e != nil {
			log.ZapLog.Info("stream response", zap.Any("err", e.(error)))
		}
	}()
	for {
		select {
		case sd := <-gs.GrpcSendClientData:
			//业务代码
			if err := stream.Send(sd); err != nil {
				log.ZapLog.Info("", zap.Any("发给网关失败err", err))
			} else {
				log.ZapLog.Info("", zap.Any("发给网关成功msg", sd.String()))
			}
		case <-stop:
			return
		}
	}
}

func (gs *streamServer) dispatch(stop chan struct{}) {
	defer func() {
		if e := recover(); e != nil {
			log.ZapLog.Info("stream dispatch", zap.Any("err", e.(error)))
		}
	}()
	for {
		select {
		case cmsg := <-gs.GrpcRecvClientData:
			{
				log.ZapLog.Info("dispatch", zap.Any("Msg", cmsg.Msg))
				if !strings.Contains(enum.CMDS, fmt.Sprintf("%d", cmsg.Msg)) {
					log.ZapLog.Error("不存在的消息", zap.Any("msg", cmsg.Msg))
				}
				if err := gs.handlerMsg(cmsg); err != nil {
					log.ZapLog.Info("handlerMsg error ", zap.Any("Msg", cmsg.Msg), zap.Any("err", err))
				}
			}
		case <-stop:
			return
		}
	}
}

// 消息处理
func (gs *streamServer) handlerMsg(clientMsg *pb.StreamRequestData) error {
	// 进入游戏，处理
	if uint16(clientMsg.Msg) == enum.CMD_ENTER_GAME_3 {
		gs.enterGame(clientMsg)
	}
	if uint16(clientMsg.Msg) == enum.CMD_INIT_GAME_3 {
		gs.initGame(clientMsg)
	}

	if uint16(clientMsg.Msg) == enum.CMD_GAME_3_BETS {
		gs.playerBet(clientMsg)
	}

	return nil
}

// ENTER_GAME 进入游戏
func (gs *streamServer) enterGame(msg *pb.StreamRequestData) error {
	var l = protoStructure.MGame_3YxxEnterGameTos{}
	if err := proto.Unmarshal(msg.Data, &l); err != nil {
		log.ZapLog.With(zap.Any("err", err)).Info("SendGameSessionInfo")
		return errors.New("proto3解码错误")
	}

	if l.GetGameId() != 3 {
		log.ZapLog.With(zap.Any("err", errors.New("GameId not match"))).Info("SendGameSessionInfo")
		return errors.New("GameId not match")
	}
	res := &protoStructure.MGame_3YxxEnterGameToc{
		GameId: l.GameId,
		Room:   l.Room,
		Desk:   l.Desk,
	}
	data, _ := proto.Marshal(res)
	sendCMsg := pb.StreamResponseData{
		ClientId: msg.GetClientId(),
		BAllUser: false,
		Uids:     nil,
		Msg:      uint32(enum.CMD_ENTER_GAME_3),
		Data:     data,
	}

	gs.GrpcSendClientData <- &sendCMsg
	return nil
}
func (gs *streamServer) initGame(msg *pb.StreamRequestData) error {
	var l = protoStructure.MInitGame_3YxxTos{}
	if err := proto.Unmarshal(msg.Data, &l); err != nil {
		log.ZapLog.With(zap.Any("err", err)).Info("SendGameSessionInfo")
		return errors.New("proto3解码错误")
	}
	// -----------------------获取玩家数据---------------------------------------
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client := GoClient()
	request := cproto.UserRequest{Uid: *l.SId}
	res, err := client.GetUserInfo(ctx, &request)
	if err != nil {
		log.ZapLog.With(zap.Error(err)).Error("grpc dial result")
	}
	// ------------------------------------------------------------------------
	playerInfo := &enum.UserInfo{
		SId:      res.GetUser().SId,
		Name:     *res.GetUser().Name,
		Sex:      *res.GetUser().Sex,
		Nickname: *res.GetUser().Nickname,
		Platform: *res.GetUser().Platform,
		Agent:    *res.GetUser().Agent,
		Coin:     *res.GetUserInfo().Gold,
		MyBet:    nil,
	}

	c.YxxGame.YxxLock.Lock()
	c.YxxGame.PlayerMap[l.GetSId()] = playerInfo
	c.YxxGame.YxxLock.Unlock()
	// -----------------------获取桌子数据----------------------------------------
	c.YxxGame.YxxLock.Lock()
	Info := c.YxxGame
	c.YxxGame.YxxLock.Unlock()
	// -----------------------------------------------------------------------

	// ----------------------------赔率信息------------------------------------
	var ood []*protoStructure.PDeskOod
	for k, v := range Info.Ood {
		tmpk := int32(k)
		tmpv := int32(v)
		tmpood := protoStructure.PDeskOod{
			Times: &tmpk,
			Ood:   &tmpv,
		}
		ood = append(ood, &tmpood)
	}
	// -----------------------------------------------------------------------

	// ----------------------------投注区域信息---------------------------------
	var areaInfo []*protoStructure.PAllAreaInfo
	for k, v := range Info.AllBet {
		tmpk := k
		tmpv := v

		tmpAreaInfo := protoStructure.PAllAreaInfo{
			Area:   &tmpk,
			AllBet: &tmpv,
			MyBet:  nil,
		}
		areaInfo = append(areaInfo, &tmpAreaInfo)
	}
	// -----------------------------------------------------------------------
	result := &protoStructure.MInitGame_3YxxToc{
		SId:       l.SId,
		Nickname:  res.GetUser().Nickname,
		Coin:      res.GetUserInfo().Gold,
		YxxOod:    ood,
		AreaInfo:  areaInfo,
		StatusNow: &Info.Status,
		NextTime:  &Info.NextTime,
		History:   nil,
		Set:       &Info.Set,
	}

	data, _ := proto.Marshal(result)
	sendCMsg := pb.StreamResponseData{
		ClientId: msg.GetClientId(),
		BAllUser: false,
		Uids:     nil,
		Msg:      uint32(enum.CMD_INIT_GAME_3),
		Data:     data,
	}
	gs.GrpcSendClientData <- &sendCMsg

	return nil
}

// PLAYER_BET   玩家下注
func (gs *streamServer) playerBet(msg *pb.StreamRequestData) error {
	var tmpG int64
	// 获取msg
	var l = protoStructure.MPlayerBetYxx_3Tos{}
	if err := proto.Unmarshal(msg.Data, &l); err != nil {
		log.ZapLog.With(zap.Any("err", err)).Info("SendGameSessionInfo")
		return errors.New("proto3解码错误")
	}

	var playerBets []*protoStructure.PPlayerBetYxx_3
	// 桌子数据
	c.YxxGame.YxxLock.Lock()
	pl := c.YxxGame.PlayerMap[l.GetSId()]

	if _, ok := pl.MyBet[l.GetArea()]; ok {
		x := c.YxxGame.Set // 局数
		t := uint32(c.YxxGame.DeskId)
		g := uint32(3) // gameId
		// ------------------------------获取玩家数据-----------------------------------------
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client := GoClient()
		request := cproto.ChangeBalanceReq{Uid: *l.SId, Gold: l.GetNums(), ChangeType: uint32(def.CHANGE_BET), PerRoundSid: &x, GameId: &g, RoomId: &t, SerialNo: nil}
		res, err := client.ReduceBalance(ctx, &request)
		if err != nil {
			log.ZapLog.With(zap.Error(err)).Error("grpc dial result")
			// TODO  给客户端发一条消息
		}

		// 下注失败
		if res.Code != 0 {
			log.ZapLog.With(zap.Error(errors.New(res.Msg))).Error("grpc dial result")
		}
		// ---------------------------------------------------------------------------------

		pl.Coin = pl.Coin - l.GetNums() // 减去玩家内存金币数量
		tmpG = pl.Coin

		if pl.MyBet != nil {
			pl.MyBet[l.GetArea()] = pl.MyBet[l.GetArea()] + l.GetNums()
		}

		if pl.MyBet == nil {
			pl.MyBet = make(map[int32]int64)
			pl.MyBet[l.GetArea()] = l.GetNums()

		}
	}

	for area, value := range pl.MyBet {
		ta := area
		tv := value

		v := protoStructure.PPlayerBetYxx_3{
			Area:      &ta,
			MyAllChip: &tv,
		}

		playerBets = append(playerBets, &v)
	}

	c.YxxGame.YxxLock.Unlock()

	betInfo := protoStructure.MPlayerBetYxx_3Toc{
		Area:      l.Area,
		Chip:      l.Nums,
		MyChip:    &tmpG,
		PlayerBet: playerBets,
	}
	// 发送msg
	data, _ := proto.Marshal(&betInfo)
	sendCMsg := pb.StreamResponseData{
		ClientId: msg.GetClientId(),
		BAllUser: false,
		Uids:     nil,
		Msg:      uint32(enum.CMD_GAME_3_BETS),
		Data:     data,
	}

	gs.GrpcSendClientData <- &sendCMsg
	return nil
}

func Run() {
	//streamIp := viper.Vp.GetString("ser.stream.ip")
	//streamPort := viper.Vp.GetInt("ser.stream.port")
	var server pb.ForwardMsgServer
	sImpl := NewStreamServer()

	server = sImpl

	g := grpc.NewServer()

	// 2.注册逻辑到server中
	pb.RegisterForwardMsgServer(g, server)

	scfg := config.NewServerCfg()
	instance := fmt.Sprintf("%s:%d", scfg.GetIp(), scfg.GetPort())

	log.ZapLog.With(zap.Any("addr", instance)).Info("Run")
	// 3.启动server
	lis, err := net.Listen("tcp", instance)
	if err != nil {
		panic("监听错误:" + err.Error())
	}

	err = g.Serve(lis)
	if err != nil {
		panic("启动错误:" + err.Error())
	}

	//sImpl.dispatch()
}

func Check() {
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for {
			<-ticker.C
			c.YxxGame.YxxLock.Lock()
			if c.YxxGame.NextTime == 1 {
				if c.YxxGame.Status == enum.Bet_Before { //
					sendToClientChangeStatus(enum.Betting, int32(10), c.YxxGame.PlayerMap)
				} else if c.YxxGame.Status == enum.Betting { //
					sendToClientChangeStatus(enum.Show_Result, int32(8), c.YxxGame.PlayerMap)
				} else if c.YxxGame.Status == enum.Show_Result { //
					sendToClientChangeStatus(enum.Bet_Before, int32(5), c.YxxGame.PlayerMap)
				}
			}
			c.YxxGame.YxxLock.Unlock()
		}
	}()
}

// sendToClientChangeStatus  告知客户端状态改变
func sendToClientChangeStatus(nextStatus, time int32, playerList map[string]*enum.UserInfo) {
	//var robotBets []*protoStructure.PRobotBet
	//// 给机器人添加筹码
	//if nextStatus == enum.Betting {
	//	var r int32 = 0
	//	var b int32 = 2
	//	var l int32 = 1
	//	var n int64 = 500000
	//	robotBets = append(robotBets, &protoStructure.PRobotBet{
	//		Area: &r,
	//		Nums: &n,
	//	}, &protoStructure.PRobotBet{
	//		Area: &b,
	//		Nums: &n,
	//	}, &protoStructure.PRobotBet{
	//		Area: &l,
	//		Nums: &n,
	//	})
	//
	//}

	var sIds []string
	for _, v := range playerList {
		sIds = append(sIds, v.SId)
	}

	// --------------------------下个状态为结算---------------------------------------
	var res []*protoStructure.PPlayerResult
	if nextStatus == enum.Show_Result {
		for _, v := range sIds {
			tmpv := v
			if info, ok := playerList[v]; ok {
				cards := model.CheckBet(c.YxxGame.Dice)
				nums := model.Settlement(info.MyBet, cards, c.YxxGame.Ood)
				dice := model.Convert(c.YxxGame.Dice)
				res = append(res, &protoStructure.PPlayerResult{
					SId:  &tmpv,
					Nums: &nums,
					Dice: dice,
				})
			}
		}
	}
	// ----------------------------------------------------------------------------

	var betInfo []*protoStructure.PAllAreaInfo
	// ----------------------------下个状态为下注--------------------------------
	// ----------------------------------------------------------------------------

	statusChange := &protoStructure.MGameStatusChangeYxx_3Toc{
		NextStatus: &nextStatus,
		Time:       &time,
		Res:        res,
		BetInfo:    betInfo,
	}

	data, _ := proto.Marshal(statusChange)
	sendCMsg := pb.StreamResponseData{
		ClientId: "",
		BAllUser: false,
		Uids:     sIds,
		Msg:      uint32(enum.CMD_GAME_3_STATUS_CHANGE),
		Data:     data,
	}

	stream.GrpcSendClientData <- &sendCMsg

}
