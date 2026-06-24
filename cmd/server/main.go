package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/dangngochoainam/gopkg/loghelper"
	"github.com/dangngochoainam/codebase-go-grpc/configs"
	"github.com/dangngochoainam/codebase-go-grpc/internal/diregistry"
	"github.com/dangngochoainam/codebase-go-grpc/pb"
	"github.com/dangngochoainam/gopkg/commonhelper"
	"github.com/dangngochoainam/gopkg/serverhelper"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func main() {

	diregistry.BuildContainer()

	config := diregistry.GetDependency[*configs.Config]()

	api := diregistry.GetDependency[pb.APIServer]()

	err := loghelper.InitZapLogger(config.Env, "codebase-go-grpc")
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, commonhelper.TRACE_ID, "123123")
	loghelper.Logger.WithContext(ctx).Info("Init logger success")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create GRPC SERVER
	grpcServer := serverhelper.CreateGrpcServer()
	pb.RegisterAPIServer(grpcServer, api)
	log.Printf("grpc-server listening at %v", lis.Addr())
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Create a client connection to the gRPC server we just started
	conn, err := grpc.Dial(fmt.Sprintf(":%s", config.GrpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	defer conn.Close()

	// Setup HTTP SERVER
	listenerHttp, err := net.Listen("tcp", fmt.Sprintf(":%s", config.HttpPort))
	if err != nil {
		log.Panic("Http Server: Can not create listener - err: ", err)
	}
	log.Printf("http-server listening at %v", listenerHttp.Addr())

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			return runtime.DefaultHeaderMatcher(key)
		}),
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			if key == "set-cookie" {
				return "Set-Cookie", true
			}
			if key == "location" {
				return "Location", true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
		runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
			// Get server metadata which contains headers set via grpc.SetHeader()
			md, ok := runtime.ServerMetadataFromContext(ctx)
			if !ok {
				return nil
			}
			// Check for location header in the outgoing metadata
			location := md.HeaderMD.Get("location")
			if len(location) > 0 {
				w.Header().Set("Location", location[0])
				w.WriteHeader(http.StatusSeeOther)
				return nil
			}
			return nil
		}),
	)

	err = pb.RegisterAPIHandler(context.Background(), mux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Cookie"},
		ExposedHeaders:   []string{"Set-Cookie"},
		// Enable Debugging for testing, consider disabling in production
		// Debug: true,
	})

	httpServer := &http.Server{
		Handler: c.Handler(mux),
	}

	log.Println("********** RUNNING **********")
	go func() {
		err := httpServer.Serve(listenerHttp)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	serverhelper.HandleGracefulShutdownServer(grpcServer, httpServer)

	// AMQP closes AFTER gRPC/HTTP have drained — no in-flight handler can race the close.
	log.Println("********** SHUTDOWN **********")
}
