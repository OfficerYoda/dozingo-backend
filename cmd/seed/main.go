package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/generated"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	if err := seed(pool); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	log.Println("Seeding completed successfully")
}

func seed(pool *pgxpool.Pool) error {
	ctx := context.Background()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback is a no-op after commit

	q := generated.New(tx)

	if err := truncateAll(ctx, tx); err != nil {
		return err
	}

	userIDs, err := seedUsers(ctx, q)
	if err != nil {
		return err
	}

	lecturerIDs, err := seedLecturers(ctx, q)
	if err != nil {
		return err
	}

	boardIDs, err := seedBoards(ctx, q, userIDs, lecturerIDs)
	if err != nil {
		return err
	}

	if err := seedCells(ctx, q, boardIDs); err != nil {
		return err
	}

	if err := seedVotes(ctx, q, userIDs, boardIDs); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// truncateAll removes all data from tables in the correct order (respecting foreign keys).
func truncateAll(ctx context.Context, tx pgx.Tx) error {
	log.Println("Truncating all tables...")
	_, err := tx.Exec(ctx, "TRUNCATE votes, cells, boards, user_authentications, lecturers, users CASCADE")
	if err != nil {
		return fmt.Errorf("truncating tables: %w", err)
	}
	return nil
}

func seedUsers(ctx context.Context, q *generated.Queries) ([]pgtype.UUID, error) {
	log.Printf("Seeding %d users...", len(users))

	ids := make([]pgtype.UUID, 0, len(users))
	for _, u := range users {
		user, err := q.CreateUser(ctx, generated.CreateUserParams{
			Username: u.Username,
			Email:    u.Email,
		})
		if err != nil {
			return nil, fmt.Errorf("creating user %q: %w", u.Username, err)
		}
		ids = append(ids, user.ID)
	}

	return ids, nil
}

func seedLecturers(ctx context.Context, q *generated.Queries) ([]pgtype.UUID, error) {
	log.Printf("Seeding %d lecturers...", len(lecturers))

	ids := make([]pgtype.UUID, 0, len(lecturers))
	for _, l := range lecturers {
		lecturer, err := q.CreateLecturer(ctx, generated.CreateLecturerParams{
			Name: l.Name,
			Slug: l.Slug,
		})
		if err != nil {
			return nil, fmt.Errorf("creating lecturer %q: %w", l.Name, err)
		}
		ids = append(ids, lecturer.ID)
	}

	return ids, nil
}

func seedBoards(ctx context.Context, q *generated.Queries, userIDs, lecturerIDs []pgtype.UUID) ([]pgtype.UUID, error) {
	log.Printf("Seeding %d boards...", len(boards))

	ids := make([]pgtype.UUID, 0, len(boards))
	for _, b := range boards {
		board, err := q.CreateBoard(ctx, generated.CreateBoardParams{
			Title:      b.Title,
			Size:       b.Size,
			AuthorID:   userIDs[b.AuthorIdx],
			LecturerID: lecturerIDs[b.LecturerIdx],
		})
		if err != nil {
			return nil, fmt.Errorf("creating board %q: %w", b.Title, err)
		}
		ids = append(ids, board.ID)
	}

	return ids, nil
}

func seedCells(ctx context.Context, q *generated.Queries, boardIDs []pgtype.UUID) error {
	totalCells := 0
	for i, b := range boards {
		phrases, ok := cellPhrases[i]
		if !ok {
			return fmt.Errorf("no cell phrases defined for board %d (%q)", i, b.Title)
		}

		for _, phrase := range phrases {
			_, err := q.CreateCell(ctx, generated.CreateCellParams{
				BoardID: boardIDs[i],
				Content: phrase,
			})
			if err != nil {
				return fmt.Errorf("creating cell for board %q: %w", b.Title, err)
			}
			totalCells++
		}
	}

	log.Printf("Seeded %d cells across %d boards", totalCells, len(boards))
	return nil
}

func seedVotes(ctx context.Context, q *generated.Queries, userIDs, boardIDs []pgtype.UUID) error {
	log.Printf("Seeding %d votes...", len(votes))

	for _, v := range votes {
		_, err := q.UpsertVote(ctx, generated.UpsertVoteParams{
			UserID:    userIDs[v.UserIdx],
			BoardID:   boardIDs[v.BoardIdx],
			VoteValue: v.Value,
		})
		if err != nil {
			return fmt.Errorf("creating vote (user=%d, board=%d): %w", v.UserIdx, v.BoardIdx, err)
		}
	}

	return nil
}
