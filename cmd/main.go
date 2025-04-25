package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcAdapter "pingpong/adapters/grpc"
	"pingpong/adapters/mysql"
	"pingpong/proto"
	"pingpong/service"
)

const (
	PlayersPort = "8888"
	TablePort   = "8889"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("🚀 Starting PingPong Bot Application with gRPC")
	log.Println("📝 Creating log file")

	f, err := os.Create("match_log.csv")
	if err != nil {
		log.Fatalf("❌ Failed to create log file: %v", err)
	}
	f.WriteString("time,turn_number,player,ball_power,routine_id,match_number\n")
	f.Close()

	log.Println("🔌 Connecting to MySQL database...")
	mysqlConfig := "root:@tcp(127.0.0.1:3306)/pingpong?parseTime=true"
	repo, err := mysql.NewMySQLRepository(mysqlConfig)
	if err != nil {
		log.Printf("⚠️ Database connection issue: %v", err)
	}
	matchService := service.NewMatchService(repo)

	playerServer := grpcAdapter.NewPlayerServer(matchService, nil)
	tableServer := grpcAdapter.NewTableServer(nil)

	go grpcAdapter.StartGRPCServer(playerServer, PlayersPort)
	go grpcAdapter.StartGRPCServer(tableServer, TablePort)

	log.Println("⏳ Waiting for gRPC servers to start...")
	time.Sleep(2 * time.Second)

	playerAddr := "localhost:" + PlayersPort
	tableAddr := "localhost:" + TablePort

	playerConn, err := grpc.Dial(playerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Failed to connect to Player service: %v", err)
	}
	defer playerConn.Close()

	tableConn, err := grpc.Dial(tableAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Failed to connect to Table service: %v", err)
	}
	defer tableConn.Close()

	playerServer.TableClient = grpcAdapter.NewTableClientFromConn(tableConn)
	tableServer.PlayerClient = grpcAdapter.NewPlayerClientFromConn(playerConn)

	log.Println("✅ Services started successfully")
	log.Println("🎮 Press Space Bar to start a new match, F2 to test DB, ESC to exit")

	go func() {
		if err := keyboard.Open(); err != nil {
			log.Fatalf("❌ Failed to open keyboard: %v", err)
		}
		defer keyboard.Close()

		playerClient := proto.NewPlayerServiceClient(playerConn)

		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				log.Printf("⚠️ Error reading key: %v", err)
				continue
			}

			switch key {
			case keyboard.KeySpace:
				log.Println("🏓 Space Bar pressed - Starting a new match...")
				_, err := playerClient.StartNewMatch(context.Background(), &proto.NewMatchRequest{})
				if err != nil {
					log.Printf("❌ Failed to start match: %v", err)
				} else {
					log.Println("✅ Match started successfully")
				}

			case keyboard.KeyF2:
				log.Println("🧪 F2 pressed - Testing DB...")
				_, err := playerClient.TestDB(context.Background(), &proto.TestDBRequest{})
				if err != nil {
					log.Printf("❌ TestDB failed: %v", err)
				} else {
					log.Println("✅ TestDB successful")
				}

			case keyboard.KeyEsc:
				log.Println("👋 ESC pressed - Exiting...")
				os.Exit(0)

			default:
				if char != 0 {
					log.Printf("🔘 Key pressed: %q", char)
				}
			}
		}
	}()

	select {}
}
