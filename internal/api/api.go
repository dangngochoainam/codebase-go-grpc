package api

import (
	"context"
	"log"

	"github.com/dangngochoainam/codebase-go-grpc/configs"
	"github.com/dangngochoainam/codebase-go-grpc/internal/usecase"
	"github.com/dangngochoainam/codebase-go-grpc/pb"
	"github.com/dangngochoainam/gopkg/copyhelper"
	"github.com/dangngochoainam/gopkg/loghelper"
	"google.golang.org/protobuf/types/known/emptypb"
)

type api struct {
	pb.UnimplementedAPIServer
	config         *configs.Config
	pbConverter    copyhelper.PbConverter
	productUsecase usecase.Product
}

func NewAPI(pbConverter copyhelper.PbConverter, config *configs.Config, productUsecase usecase.Product) pb.APIServer {
	return &api{
		config:         config,
		pbConverter:    pbConverter,
		productUsecase: productUsecase,
	}
}

func (s *api) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *api) HealthCheck(ctx context.Context, in *emptypb.Empty) (*pb.HelloReply, error) {
	loghelper.Logger.WithContext(ctx).Info("HealthCheck")
	return &pb.HelloReply{Message: "OK"}, nil
}
