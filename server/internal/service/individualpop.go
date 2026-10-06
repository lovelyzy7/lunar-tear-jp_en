package service

import (
	"context"
	"log"

	pb "lunar-tear/server/gen/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

// IndividualpopServiceServer 返回"未读个性化弹窗"列表。
// 私服不需要首页/活动弹窗，直接返回空列表即可：
// 主要目的是避免客户端在启动流程里拿到 gRPC Unimplemented 错误。
type IndividualpopServiceServer struct {
	pb.UnimplementedIndividualpopServiceServer
}

func NewIndividualpopServiceServer() *IndividualpopServiceServer {
	return &IndividualpopServiceServer{}
}

func (s *IndividualpopServiceServer) GetUnreadPop(ctx context.Context, _ *emptypb.Empty) (*pb.GetUnreadPopResponse, error) {
	log.Printf("[IndividualpopService] GetUnreadPop -> empty stub")
	return &pb.GetUnreadPopResponse{}, nil
}
