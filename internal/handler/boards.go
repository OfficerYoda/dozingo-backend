package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

/// ===== Input/Output types =====

type BoardOutput struct {
	ID         string `json:"id" format:"uuid"`
	Title      string `json:"title" format:"text"`
	Size       int32  `json:"size" format:"integer"`
	AuthorID   string `json:"author_id" format:"uuid"`
	LecturerID string `json:"lecturer_id" format:"uuid"`
}

type GetBoardsInput struct {
	AuthorID   string `query:"author_id"`
	LecturerID string `query:"lecturer_id"`
	Size       int    `query:"size"`
}

type GetBoardsOutput struct {
	Body []BoardOutput `json:"boards"`
}

type GetBoardByIDInput struct {
	ID string `path:"id" format:"uuid"`
}

type GetBoardByIDOutput struct {
	Body BoardOutput
}

type CreateBoardInput struct {
	Body struct {
		Title      string `json:"title" format:"text" required:"true" maxLength:"200"`
		Size       int32  `json:"size" format:"integer" required:"true" maxLength:"200"`
		AuthorID   string `json:"author_id" format:"uuid" required:"true"`
		LecturerID string `json:"lecturer_id" format:"uuid" required:"false"`
	}
}

type CreateBoardOutput struct {
	Body BoardOutput
}

type DeleteBoardInput struct {
	ID string `path:"id" format:"uuid"`
}

/// ===== Register =====

func RegisterBoards(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	_ = queries

	huma.Register(api, huma.Operation{
		OperationID: "get-boards",
		Method:      http.MethodGet,
		Path:        "/boards",
		Summary:     "Get all boards with optional filters",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *GetBoardsInput) (*GetBoardsOutput, error) {
		return getBoards(ctx, pool, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-board-by-id",
		Method:      http.MethodGet,
		Path:        "/boards/{id}",
		Summary:     "Get a board by ID",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *GetBoardByIDInput) (*GetBoardByIDOutput, error) {
		return getBoardByID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-board",
		Method:      http.MethodPost,
		Path:        "/boards",
		Summary:     "Create a board",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *CreateBoardInput) (*CreateBoardOutput, error) {
		return createBoard(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-board",
		Method:      http.MethodDelete,
		Path:        "/boards/{id}",
		Summary:     "Delete a board",
		Tags:        []string{"Boards"},
	}, func(ctx context.Context, input *DeleteBoardInput) (*struct{}, error) {
		return deleteBoard(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getBoards(ctx context.Context, pool *pgxpool.Pool, input GetBoardsInput) (*GetBoardsOutput, error) {
	rows, err := queryBoardsFiltered(ctx, input, pool)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to query boards", err)
	}
	defer rows.Close()

	boards, err := pgx.CollectRows(rows, pgx.RowToStructByName[generated.Board])
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to scan boards", err)
	}

	out := &GetBoardsOutput{}
	out.Body = make([]BoardOutput, 0)
	for _, b := range boards {
		out.Body = append(out.Body, boardToOutput(b))
	}

	return out, nil
}

func getBoardByID(ctx context.Context, queries *generated.Queries, input GetBoardByIDInput) (*GetBoardByIDOutput, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id", err)
	}

	board, err := queries.GetBoardByID(ctx, id)
	if err != nil {
		return nil, huma.Error404NotFound("board not found", err)
	}

	return &GetBoardByIDOutput{Body: boardToOutput(board)}, nil
}

func createBoard(ctx context.Context, queries *generated.Queries, input CreateBoardInput) (*CreateBoardOutput, error) {
	authorID, err := uuidFromString(input.Body.AuthorID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid author_id", err)
	}
	lecturerID, err := uuidFromString(input.Body.LecturerID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid lecturer_id", err)
	}

	board, err := queries.CreateBoard(ctx, generated.CreateBoardParams{
		Title:      input.Body.Title,
		Size:       input.Body.Size,
		AuthorID:   authorID,
		LecturerID: lecturerID,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create board", err)
	}

	return &CreateBoardOutput{Body: boardToOutput(board)}, nil
}

func deleteBoard(ctx context.Context, queries *generated.Queries, input DeleteBoardInput) (*struct{}, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id", err)
	}

	_, err = queries.DeleteBoard(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("board not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to delete board", err)
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func queryBoardsFiltered(ctx context.Context, input GetBoardsInput, pool *pgxpool.Pool) (pgx.Rows, error) {
	query := "SELECT * FROM boards WHERE 1=1"
	args := []any{}
	i := 1

	if input.AuthorID != "" {
		query += fmt.Sprintf(" AND author_id = $%d", i)
		args = append(args, input.AuthorID)
		i++
	}

	if input.LecturerID != "" {
		query += fmt.Sprintf(" AND lecturer_id = $%d", i)
		args = append(args, input.LecturerID)
		i++
	}

	if input.Size != 0 {
		query += fmt.Sprintf(" AND size = $%d", i)
		args = append(args, input.Size)
	}

	query += " ORDER BY created_at DESC;"

	rows, err := pool.Query(ctx, query, args...)

	return rows, err
}

func boardToOutput(board generated.Board) BoardOutput {
	return BoardOutput{
		ID:         board.ID.String(),
		Title:      board.Title,
		Size:       board.Size,
		AuthorID:   board.AuthorID.String(),
		LecturerID: board.LecturerID.String(),
	}
}
