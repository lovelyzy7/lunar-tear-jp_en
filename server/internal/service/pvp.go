package service

import (
	"context"
	"log"

	pb "lunar-tear/server/gen/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

// PvpServiceServer 是竞技场的空实现（stub）。
//
// 全部接口返回空响应，目的是：
//  1. 避免客户端在首页/竞技场入口请求时收到 gRPC Unimplemented 报错；
//  2. 竞技场本身仍不可游玩（无匹配、无排名、无对战数据）。
//
// 如果不需要这个 stub，删除本文件与 grpc.go 里的注册即可。
type PvpServiceServer struct {
	pb.UnimplementedPvpServiceServer
}

func NewPvpServiceServer() *PvpServiceServer {
	return &PvpServiceServer{}
}

func (s *PvpServiceServer) GetTopData(context.Context, *emptypb.Empty) (*pb.GetTopDataResponse, error) {
	log.Printf("[PvpService] GetTopData -> empty stub")
	return &pb.GetTopDataResponse{}, nil
}

func (s *PvpServiceServer) GetMatchingList(context.Context, *emptypb.Empty) (*pb.GetMatchingListResponse, error) {
	log.Printf("[PvpService] GetMatchingList -> empty stub")
	return &pb.GetMatchingListResponse{}, nil
}

func (s *PvpServiceServer) UpdateMatchingList(context.Context, *emptypb.Empty) (*pb.UpdateMatchingListResponse, error) {
	log.Printf("[PvpService] UpdateMatchingList -> empty stub")
	return &pb.UpdateMatchingListResponse{}, nil
}

func (s *PvpServiceServer) StartBattle(_ context.Context, _ *pb.StartBattleRequest) (*pb.StartBattleResponse, error) {
	log.Printf("[PvpService] StartBattle -> empty stub")
	return &pb.StartBattleResponse{}, nil
}

func (s *PvpServiceServer) FinishBattle(_ context.Context, _ *pb.FinishBattleRequest) (*pb.FinishBattleResponse, error) {
	log.Printf("[PvpService] FinishBattle -> empty stub")
	return &pb.FinishBattleResponse{}, nil
}

func (s *PvpServiceServer) GetRanking(_ context.Context, _ *pb.GetRankingRequest) (*pb.GetRankingResponse, error) {
	log.Printf("[PvpService] GetRanking -> empty stub")
	return &pb.GetRankingResponse{}, nil
}

func (s *PvpServiceServer) GetSeasonResult(context.Context, *emptypb.Empty) (*pb.GetSeasonResultResponse, error) {
	log.Printf("[PvpService] GetSeasonResult -> empty stub")
	return &pb.GetSeasonResultResponse{}, nil
}

func (s *PvpServiceServer) GetAttackLogList(context.Context, *emptypb.Empty) (*pb.GetAttackLogListResponse, error) {
	log.Printf("[PvpService] GetAttackLogList -> empty stub")
	return &pb.GetAttackLogListResponse{}, nil
}

func (s *PvpServiceServer) GetDefenseLogList(context.Context, *emptypb.Empty) (*pb.GetDefenseLogListResponse, error) {
	log.Printf("[PvpService] GetDefenseLogList -> empty stub")
	return &pb.GetDefenseLogListResponse{}, nil
}
