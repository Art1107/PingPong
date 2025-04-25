package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/emptypb"

	"pingpong/domain"
	"pingpong/ports"
	pb "pingpong/proto"
)

const (
	PlayersPort = "8888"
	TablePort   = "8889"
)

type PlayerServer struct {
	pb.UnimplementedPlayerServiceServer
	matchService     ports.MatchService
	currentMatch     domain.Match
	turnCounter      int
	matchNumberCount int
	routineID        string
	matchesMutex     sync.Mutex
	TableClient      pb.TableServiceClient
	gameActive       bool
}

func NewPlayerServer(matchService ports.MatchService, tableConn *grpc.ClientConn) *PlayerServer {
	return &PlayerServer{
		matchService: matchService,
		TableClient:  pb.NewTableServiceClient(tableConn),
	}
}

func (s *PlayerServer) initMatch() {
	s.matchNumberCount++
	s.currentMatch = domain.Match{
		ID:          0,
		MatchNumber: s.matchNumberCount,
		StartTime:   time.Now(),
		Turns:       []domain.Turn{},
	}
	s.turnCounter = 0
	s.routineID = fmt.Sprintf("match-%d-%s", s.matchNumberCount, time.Now().Format("20060102150405"))
	s.gameActive = true
	log.Printf("🆕 New match initialized: Match #%d, RoutineID: %s", s.matchNumberCount, s.routineID)
}

func (s *PlayerServer) logTurn(player string, ballPower int) {
	turn := domain.Turn{
		TurnNumber:  s.turnCounter,
		Time:        time.Now(),
		Player:      player,
		BallPower:   ballPower,
		RoutineID:   s.routineID,
		MatchNumber: s.currentMatch.MatchNumber,
	}

	s.matchesMutex.Lock()
	s.currentMatch.Turns = append(s.currentMatch.Turns, turn)
	s.matchesMutex.Unlock()

	log.Printf("🏓 Turn #%d: Player %s hit with power %d (Match #%d, Routine: %s)",
		turn.TurnNumber, player, ballPower, turn.MatchNumber, s.routineID)

	f, err := os.OpenFile("match_log.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("❌ Error opening log file: %v", err)
		return
	}
	defer f.Close()

	logLine := fmt.Sprintf("%s,%d,%s,%d,%s,%d\n",
		turn.Time.Format(time.RFC3339),
		turn.TurnNumber,
		turn.Player,
		turn.BallPower,
		turn.RoutineID,
		turn.MatchNumber)

	if _, err := f.WriteString(logLine); err != nil {
		log.Printf("❌ Error writing to log file: %v", err)
	}
}

func (s *PlayerServer) StartNewMatch(ctx context.Context, req *pb.NewMatchRequest) (*pb.NewMatchResponse, error) {
	log.Println("🎮 Starting new match...")
	s.initMatch()

	go func() {
		time.Sleep(100 * time.Millisecond)
		log.Printf("📤 Sending start game request to table")

		_, err := s.TableClient.StartGame(context.Background(), &pb.StartGameRequest{})
		if err != nil {
			log.Printf("❌ Failed to notify Table: %v", err)
		} else {
			log.Printf("✅ Successfully notified Table")
		}
	}()

	return &pb.NewMatchResponse{Message: "New match started"}, nil
}

func (s *PlayerServer) PlayerAPing(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	log.Println("📥 Player A received ping")

	s.turnCounter++
	receivedPower := int(req.BallPower)
	s.logTurn("A", receivedPower)

	returnPower := int(float64(receivedPower) * (70 + float64(time.Now().UnixNano()%20)) / 100)
	log.Printf("🎾 Player A returning with power: %d (70-90%% of %d)", returnPower, receivedPower)

	go func() {
		log.Printf("📤 Player A sending to table with ball power: %d", returnPower)

		_, err := s.TableClient.ReceiveBall(context.Background(), &pb.ReceiveBallRequest{
			BallPower:  int32(returnPower),
			FromPlayer: "A",
		})
		if err != nil {
			log.Printf("❌ Failed to ping table: %v", err)
			return
		}
		log.Printf("✅ Successfully sent ping to table")
	}()

	return &pb.PingResponse{}, nil
}


func (s *PlayerServer) PlayerBPing(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	log.Println("📥 Player B received ping")

	if !s.gameActive {
		log.Println("🚫 Match already ended. Ignoring ping.")
		return &pb.PingResponse{}, nil
	}

	s.turnCounter++
	receivedPower := int(req.BallPower)
	s.logTurn("B", receivedPower)

	returnPower := 50 + int(time.Now().UnixNano()%50)
	log.Printf("🎾 Player B generated return power: %d (vs received power: %d)", returnPower, receivedPower)

	if s.turnCounter > 10 {
		log.Printf("⚠️ Exceeded 10 turns! Evaluating winner by power...")

		s.currentMatch.EndTime = time.Now()
		s.gameActive = false

		if returnPower > receivedPower {
			s.currentMatch.Winner = "B"
			log.Printf("🏁 Match ended by rule (10 turns). Winner: Player B")
		} else if returnPower < receivedPower {
			s.currentMatch.Winner = "A"
			log.Printf("🏁 Match ended by rule (10 turns). Winner: Player A")
		} else {
			s.currentMatch.Winner = "Draw"
			log.Printf("🤝 Match ended in a draw (equal power)")
		}

		err := s.matchService.SaveMatch(context.Background(), s.currentMatch)
		if err != nil {
			log.Printf("❌ Error saving to MySQL: %v", err)
		} else {
			log.Println("✅ Match saved to MySQL successfully")
		}

		return &pb.PingResponse{}, nil
	}

	if returnPower > receivedPower {
		log.Printf("✅ Player B returns the ball (power %d > %d)", returnPower, receivedPower)
		go func() {
			log.Printf("📤 Player B sending to table with ball power: %d", returnPower)

			_, err := s.TableClient.ReceiveBall(context.Background(), &pb.ReceiveBallRequest{
				BallPower:  int32(returnPower),
				FromPlayer: "B",
			})
			if err != nil {
				log.Printf("❌ Failed to ping table: %v", err)
				return
			}
			log.Printf("✅ Successfully sent ping to table")
		}()
	} else {
		log.Printf("❌ Player B lost! (return power %d <= received power %d)", returnPower, receivedPower)
		s.currentMatch.EndTime = time.Now()
		s.currentMatch.Winner = "A"
		s.gameActive = false

		log.Println("Saving final match result to database...")
		err := s.matchService.SaveMatch(context.Background(), s.currentMatch)
		if err != nil {
			log.Printf("❌ Error saving to MySQL: %v", err)
		} else {
			log.Println("✅ Match saved to MySQL successfully")
		}

		log.Printf("🏁 Game ended! Winner: Player A (Match #%d)", s.currentMatch.MatchNumber)
	}

	return &pb.PingResponse{}, nil
}

func (s *PlayerServer) GetMatch(ctx context.Context, req *pb.GetMatchRequest) (*pb.Match, error) {
	log.Println("📊 Request for last match")

	match, err := s.matchService.GetLastMatch(ctx)
	if err != nil {
		log.Printf("❌ No match data available: %v", err)
		return nil, fmt.Errorf("no match data available: %v", err)
	}

	pbMatch := convertDomainMatchToProto(match)
	log.Printf("✅ Found last match data")
	return pbMatch, nil
}

func (s *PlayerServer) GetMatchByID(ctx context.Context, req *pb.GetMatchByIDRequest) (*pb.Match, error) {
	id := int(req.Id)
	log.Printf("📊 Request for match with ID: %d", id)

	match, err := s.matchService.GetMatchByID(ctx, id)
	if err != nil {
		log.Printf("❌ Match not found: %v", err)
		return nil, fmt.Errorf("match not found: %v", err)
	}

	pbMatch := convertDomainMatchToProto(match)
	log.Printf("✅ Found match data for ID %d", id)
	return pbMatch, nil
}

func (s *PlayerServer) TestDB(ctx context.Context, req *pb.TestDBRequest) (*pb.TestDBResponse, error) {
	log.Println("🧪 Testing database connections...")

	err := s.matchService.TestConnection(ctx)
	if err != nil {
		log.Printf("❌ Database test failed: %v", err)
		return nil, fmt.Errorf("database test failed: %v", err)
	}

	return &pb.TestDBResponse{Message: "Database test completed successfully"}, nil
}

func (s *PlayerServer) IsGameActive(ctx context.Context, _ *emptypb.Empty) (*pb.IsGameActiveResponse, error) {
	return &pb.IsGameActiveResponse{
		Active: s.gameActive,
	}, nil
}

func convertDomainMatchToProto(match domain.Match) *pb.Match {
	pbMatch := &pb.Match{
		Id:          int32(match.ID),
		MatchNumber: int32(match.MatchNumber),
		StartTime:   timestamppb.New(match.StartTime),
		Winner:      match.Winner,
	}

	if !match.EndTime.IsZero() {
		pbMatch.EndTime = timestamppb.New(match.EndTime)
	}

	pbTurns := make([]*pb.Turn, len(match.Turns))
	for i, turn := range match.Turns {
		pbTurns[i] = &pb.Turn{
			Id:          int32(turn.ID),
			TurnNumber:  int32(turn.TurnNumber),
			Time:        timestamppb.New(turn.Time),
			Player:      turn.Player,
			BallPower:   int32(turn.BallPower),
			RoutineId:   turn.RoutineID,
			MatchNumber: int32(turn.MatchNumber),
		}
	}
	pbMatch.Turns = pbTurns

	return pbMatch
}

type TableServer struct {
	pb.UnimplementedTableServiceServer
	PlayerClient pb.PlayerServiceClient
}

func NewTableServer(playerConn *grpc.ClientConn) *TableServer {
	return &TableServer{
		PlayerClient: pb.NewPlayerServiceClient(playerConn),
	}
}

func (s *TableServer) StartGame(ctx context.Context, req *pb.StartGameRequest) (*pb.StartGameResponse, error) {
	log.Println("🎮 Table received start game request")
	
	initialPower := 70 + int(time.Now().UnixNano()%30)
	log.Printf("🎾 Starting game with initial power: %d", initialPower)
	
	go func() {
		log.Printf("📤 Table sending to Player A with initial power: %d", initialPower)
		
		_, err := s.PlayerClient.PlayerAPing(context.Background(), &pb.PingRequest{
			BallPower: int32(initialPower),
		})
		if err != nil {
			log.Printf("❌ Failed to ping Player A: %v", err)
			return
		}
		log.Printf("✅ Successfully sent initial ping to Player A")
	}()
	
	return &pb.StartGameResponse{Message: "Game started"}, nil
}

func (s *TableServer) ReceiveBall(ctx context.Context, req *pb.ReceiveBallRequest) (*pb.ReceiveBallResponse, error) {
	log.Println("📥 Table received ball")

	ballPower := int(req.BallPower)
	fromPlayer := req.FromPlayer

	log.Printf("🎾 Table received ball from Player %s with power %d", fromPlayer, ballPower)

	activeRes, err := s.PlayerClient.IsGameActive(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Printf("❌ Failed to check game status: %v", err)
	} else if !activeRes.Active {
		log.Printf("🏁 Match already ended (checked via PlayerServer). Not forwarding ball.")
		return &pb.ReceiveBallResponse{}, nil
	}	

	go func() {
		if fromPlayer == "A" {
			log.Printf("📤 Table forwarding ball to Player B")
			_, err := s.PlayerClient.PlayerBPing(context.Background(), &pb.PingRequest{
				BallPower: int32(ballPower),
			})
			if err != nil {
				log.Printf("❌ Failed to forward ball to Player B: %v", err)
				return
			}
		} else {
			log.Printf("📤 Table forwarding ball to Player A")
			_, err := s.PlayerClient.PlayerAPing(context.Background(), &pb.PingRequest{
				BallPower: int32(ballPower),
			})
			if err != nil {
				log.Printf("❌ Failed to forward ball to Player A: %v", err)
				return
			}
		}
		log.Printf("✅ Successfully forwarded ball")
	}()

	return &pb.ReceiveBallResponse{}, nil
}

func StartGRPCServer(server interface{}, port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", port, err)
	}
	
	grpcServer := grpc.NewServer()
	
	switch s := server.(type) {
	case *PlayerServer:
		pb.RegisterPlayerServiceServer(grpcServer, s)
		log.Printf("🏓 Player gRPC server starting on port %s", port)
	case *TableServer:
		pb.RegisterTableServiceServer(grpcServer, s)
		log.Printf("🏓 Table gRPC server starting on port %s", port)
	}
	
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}