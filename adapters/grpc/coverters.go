package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	
	"pingpong/domain"
	pb "pingpong/proto"
)

func DomainMatchToProto(match domain.Match) *pb.Match {
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
		pbTurns[i] = DomainTurnToProto(turn)
	}
	pbMatch.Turns = pbTurns

	return pbMatch
}

func DomainTurnToProto(turn domain.Turn) *pb.Turn {
	return &pb.Turn{
		Id:          int32(turn.ID),
		TurnNumber:  int32(turn.TurnNumber),
		Time:        timestamppb.New(turn.Time),
		Player:      turn.Player,
		BallPower:   int32(turn.BallPower),
		RoutineId:   turn.RoutineID,
		MatchNumber: int32(turn.MatchNumber),
	}
}

func ProtoToDomainMatch(pbMatch *pb.Match) domain.Match {
	match := domain.Match{
		ID:          int(pbMatch.Id),
		MatchNumber: int(pbMatch.MatchNumber),
		StartTime:   pbMatch.StartTime.AsTime(),
		Winner:      pbMatch.Winner,
	}

	if pbMatch.EndTime != nil {
		match.EndTime = pbMatch.EndTime.AsTime()
	}

	turns := make([]domain.Turn, len(pbMatch.Turns))
	for i, pbTurn := range pbMatch.Turns {
		turns[i] = ProtoToDomainTurn(pbTurn)
	}
	match.Turns = turns

	return match
}

func ProtoToDomainTurn(pbTurn *pb.Turn) domain.Turn {
	return domain.Turn{
		ID:          int(pbTurn.Id),
		TurnNumber:  int(pbTurn.TurnNumber),
		Time:        pbTurn.Time.AsTime(),
		Player:      pbTurn.Player,
		BallPower:   int(pbTurn.BallPower),
		RoutineID:   pbTurn.RoutineId,
		MatchNumber: int(pbTurn.MatchNumber),
	}
}