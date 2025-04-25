package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"pingpong/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(connectionString string) (*MySQLRepository, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 3)


	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	err = initSchema(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	return &MySQLRepository{db: db}, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS matches (
			id INT AUTO_INCREMENT PRIMARY KEY,
			match_number INT NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NULL,
			winner VARCHAR(10) NULL,
			turns JSON NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS turns (
			id INT AUTO_INCREMENT PRIMARY KEY,
			turn_number INT NOT NULL,
			time TIMESTAMP NOT NULL,
			player VARCHAR(10) NOT NULL,
			ball_power INT NOT NULL,
			routine_id VARCHAR(50) NOT NULL,
			match_number INT NOT NULL,
			match_id INT NOT NULL,
			FOREIGN KEY (match_id) REFERENCES matches(id)
		)
	`)
	return err
}

func (r *MySQLRepository) SaveMatch(ctx context.Context, match domain.Match) error {
	log.Printf("💾 Saving complete match #%d to MySQL...", match.MatchNumber)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var result sql.Result
	query := `INSERT INTO matches (match_number, start_time, end_time, winner) 
			  VALUES (?, ?, ?, ?)`
	result, err = tx.ExecContext(ctx, query, match.MatchNumber, match.StartTime, 
		match.EndTime, match.Winner)

	if err != nil {
		return fmt.Errorf("failed to save match: %v", err)
	}

	matchID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %v", err)
	}

	for _, turn := range match.Turns {
		query := `INSERT INTO turns (turn_number, time, player, ball_power, routine_id, match_number, match_id) 
				  VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.ExecContext(ctx, query, turn.TurnNumber, turn.Time, turn.Player, 
			turn.BallPower, turn.RoutineID, turn.MatchNumber, matchID)
		if err != nil {
			return fmt.Errorf("failed to save turn: %v", err)
		}
	}

	turnsJSON, err := json.Marshal(match.Turns)
	if err != nil {
		return fmt.Errorf("failed to marshal turns: %v", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE matches SET turns = ? WHERE id = ?", turnsJSON, matchID)
	if err != nil {
		return fmt.Errorf("failed to update turns JSON: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ Complete match saved to MySQL successfully with ID: %d", matchID)
	return nil
}

func (r *MySQLRepository) GetMatchByID(ctx context.Context, id int) (domain.Match, error) {
	log.Printf("📊 Fetching match with ID: %d from MySQL", id)

	var match domain.Match
	var turnsJSON []byte

	query := `SELECT id, match_number, start_time, end_time, winner, turns FROM matches WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID, &match.MatchNumber, &match.StartTime, &match.EndTime, &match.Winner, &turnsJSON)
	if err != nil {
		return domain.Match{}, fmt.Errorf("failed to get match: %v", err)
	}

	if turnsJSON != nil {
		err = json.Unmarshal(turnsJSON, &match.Turns)
		if err != nil {
			return domain.Match{}, fmt.Errorf("failed to unmarshal turns: %v", err)
		}
	} else {
		match.Turns = []domain.Turn{}
		rows, err := r.db.QueryContext(ctx, 
			`SELECT id, turn_number, time, player, ball_power, routine_id, match_number 
             FROM turns WHERE match_id = ? ORDER BY turn_number`, id)
		if err != nil {
			return domain.Match{}, fmt.Errorf("failed to fetch turns: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var turn domain.Turn
			err := rows.Scan(&turn.ID, &turn.TurnNumber, &turn.Time, &turn.Player, 
				&turn.BallPower, &turn.RoutineID, &turn.MatchNumber)
			if err != nil {
				return domain.Match{}, fmt.Errorf("failed to scan turn: %v", err)
			}
			match.Turns = append(match.Turns, turn)
		}
	}

	log.Printf("✅ Found match with ID %d containing %d turns", match.ID, len(match.Turns))
	return match, nil
}

func (r *MySQLRepository) GetLastMatch(ctx context.Context) (domain.Match, error) {
	log.Println("📊 Fetching last match from MySQL")

	var id int
	err := r.db.QueryRowContext(ctx, "SELECT MAX(id) FROM matches").Scan(&id)
	if err != nil {
		return domain.Match{}, fmt.Errorf("failed to get last match ID: %v", err)
	}

	return r.GetMatchByID(ctx, id)
}

func (r *MySQLRepository) TestConnection(ctx context.Context) error {
	log.Println("🧪 Testing MySQL connection...")
	err := r.db.PingContext(ctx)
	if err != nil {
		log.Printf("❌ Failed to ping MySQL: %v", err)
		return err
	}

	_, err = r.db.ExecContext(ctx, 
		"INSERT INTO matches (match_number, start_time) VALUES (?, ?)", 
		0, time.Now())
	if err != nil {
		log.Printf("❌ Failed to insert test record: %v", err)
		return err
	}

	log.Println("✅ MySQL connection test successful")
	return nil
}