package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	pb "pingpong/proto"
)

func NewPlayerClientFromConn(conn *grpc.ClientConn) pb.PlayerServiceClient {
	return pb.NewPlayerServiceClient(conn)
}

func StartNewMatch(client pb.PlayerServiceClient) (*pb.NewMatchResponse, error) {
	log.Println("📤 Client sending StartNewMatch request")
	return client.StartNewMatch(context.Background(), &pb.NewMatchRequest{})
}

func PlayerAPing(client pb.PlayerServiceClient, ballPower int32) (*pb.PingResponse, error) {
	log.Printf("📤 Client sending PlayerAPing request with power: %d", ballPower)
	return client.PlayerAPing(context.Background(), &pb.PingRequest{BallPower: ballPower})
}

func PlayerBPing(client pb.PlayerServiceClient, ballPower int32) (*pb.PingResponse, error) {
	log.Printf("📤 Client sending PlayerBPing request with power: %d", ballPower)
	return client.PlayerBPing(context.Background(), &pb.PingRequest{BallPower: ballPower})
}

func GetMatch(client pb.PlayerServiceClient) (*pb.Match, error) {
	log.Println("📤 Client sending GetMatch request")
	return client.GetMatch(context.Background(), &pb.GetMatchRequest{})
}

func GetMatchByID(client pb.PlayerServiceClient, id int32) (*pb.Match, error) {
	log.Printf("📤 Client sending GetMatchByID request for ID: %d", id)
	return client.GetMatchByID(context.Background(), &pb.GetMatchByIDRequest{Id: id})
}

func TestDB(client pb.PlayerServiceClient) (*pb.TestDBResponse, error) {
	log.Println("📤 Client sending TestDB request")
	return client.TestDB(context.Background(), &pb.TestDBRequest{})
}

func NewTableClientFromConn(conn *grpc.ClientConn) pb.TableServiceClient {
	return pb.NewTableServiceClient(conn)
}

func StartGame(client pb.TableServiceClient) (*pb.StartGameResponse, error) {
	log.Println("📤 Client sending StartGame request")
	return client.StartGame(context.Background(), &pb.StartGameRequest{})
}

func ReceiveBall(client pb.TableServiceClient, ballPower int32, fromPlayer string) (*pb.ReceiveBallResponse, error) {
	log.Printf("📤 Client sending ReceiveBall request: power %d from player %s", ballPower, fromPlayer)
	return client.ReceiveBall(context.Background(), &pb.ReceiveBallRequest{
		BallPower:  ballPower,
		FromPlayer: fromPlayer,
	})
}