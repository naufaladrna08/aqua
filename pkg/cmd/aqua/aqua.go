package aqua

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/naufaladrna08/aqua/pkg/api/aqua"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	service "github.com/naufaladrna08/aqua/pkg/service/aqua"
)

var (
	// command-line options
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:9090", "gRPC server endpoint")
)

type Config struct {
	HttpPort string
	GRPCPort string

	/* Database config */
	DBUsername string
	DBPassword string
	DBPort     string
	DBHost     string
	DBName     string
}

func RunServer() error {
	/* Parse the arguments */
	var cfg Config
	flag.StringVar(&cfg.HttpPort, "http-port", "", "Port untuk server HTTP")
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "Port untuk service gRPC")
	flag.StringVar(&cfg.DBPort, "db-port", "3306", "Port untuk database")
	flag.StringVar(&cfg.DBHost, "db-host", "localhost", "Host untuk database")
	flag.StringVar(&cfg.DBUsername, "db-user", "root", "User untuk database")
	flag.StringVar(&cfg.DBPassword, "db-password", "", "Password untuk database")
	flag.StringVar(&cfg.DBName, "db-schema", "", "Schema database")
	flag.Parse()

	if cfg.HttpPort == "" {
		log.Printf("Argumen HTTP port tidak tersedia, menggunakan default 8080. Gunakan -http-port\n")
		cfg.HttpPort = "8080"
	}

	if cfg.GRPCPort == "" {
		log.Printf("Argumen gRPC port tidak tersedia, menggunakan default 9090. Gunakan -grpc-port\n")
		cfg.GRPCPort = "9090"
	}

	if cfg.DBName == "" {
		log.Print("Argumen database tidak tersedia. Silahkan mengisi nama database dengan -db-schema\n")

		err := errors.New("fuck you")
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a connection to database using GORM
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
	db, dbErr := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if dbErr != nil {
		log.Printf("Failed to open database: %v\n", dbErr)
	}

	_ = db

	// Register gRPC Endpoint
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	grpcAquaService := service.NewAquaServiceServer(db)

	// HTTP
	err := aqua.RegisterAquaServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	} else {
		log.Printf("gRPC service is started at port %s\n", cfg.GRPCPort)
	}

	// Start gRPC service
	go func() {
		var opts []grpc.ServerOption
		port, _ := strconv.Atoi(cfg.GRPCPort)
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost: %d", port))

		if err != nil {
			log.Fatalf("Failed to run gRPC service: %s", err.Error())
		}

		grpcServer := grpc.NewServer(opts...)
		aqua.RegisterAquaServiceServer(grpcServer, grpcAquaService)

		grpcServer.Serve(lis)
	}()

	// Start gRPC server
	err = http.ListenAndServe(":"+cfg.HttpPort, mux)

	if err != nil {
		log.Fatalf("Faild to run HTTP server: %v", err.Error())
	}

	return err
}
