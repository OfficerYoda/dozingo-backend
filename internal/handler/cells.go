package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

/// ===== Input/Output types =====

type CellOutput struct {
	ID      string `json:"id" format:"uuid"`
	BoardID string `json:"board_id" format:"uuid"`
	Content string `json:"content" format:"text"`
}

type GetCellsByBoardIDInput struct {
	BoardID string `path:"board_id"`
}

type GetCellsByBoardIDOutput struct {
	Body []CellOutput `json:"cells"`
}

type CreateCellInput struct {
	BoardID string `path:"board_id" format:"uuid"`
	Body    struct {
		Content string `json:"content" format:"text" required:"true" maxLength:"200"`
	}
}

type CreateCellOutput struct {
	Body CellOutput
}

type UpdateCellInput struct {
	BoardID string `path:"board_id" format:"uuid"`
	CellID  string `path:"cell_id" format:"uuid"`
	Body    struct {
		Content string `json:"content" format:"text" required:"true" maxLength:"200"`
	}
}

type UpdateCellOutput struct {
	Body CellOutput `json:"cell"`
}

type DeleteCellInput struct {
	BoardID string `path:"board_id" format:"uuid"`
	CellID  string `path:"cell_id" format:"uuid"`
}

/// ===== Register =====

func RegisterCells(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	_ = queries

	huma.Register(api, huma.Operation{
		OperationID: "get-cells-by-board-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/cells",
		Summary:     "Get all cells for a Board",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *GetCellsByBoardIDInput) (*GetCellsByBoardIDOutput, error) {
		return getCellsByBoardID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-cell",
		Method:      http.MethodPost,
		Path:        "/boards/{board_id}/cells",
		Summary:     "Create a cell",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *CreateCellInput) (*CreateCellOutput, error) {
		return createCell(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-cell",
		Method:      http.MethodPut,
		Path:        "/boards/{board_id}/cells/{cell_id}",
		Summary:     "Update a cell",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *UpdateCellInput) (*UpdateCellOutput, error) {
		return updateCell(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-cell",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}/cells/{cell_id}",
		Summary:     "Delete a cell",
		Tags:        []string{"Cells"},
	}, func(ctx context.Context, input *DeleteCellInput) (*struct{}, error) {
		return deleteCell(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getCellsByBoardID(ctx context.Context, queries *generated.Queries, input GetCellsByBoardIDInput) (*GetCellsByBoardIDOutput, error) {
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	cells, err := queries.GetCellsByBoardID(ctx, boardID)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error", err)
	}

	out := &GetCellsByBoardIDOutput{}
	out.Body = make([]CellOutput, 0)
	for _, c := range cells {
		out.Body = append(out.Body, cellToOutput(c))
	}

	return out, nil
}

func createCell(ctx context.Context, queries *generated.Queries, input CreateCellInput) (*CreateCellOutput, error) {
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	board, err := queries.CreateCell(ctx, generated.CreateCellParams{
		BoardID: boardID,
		Content: input.Body.Content,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create cell", err)
	}

	return &CreateCellOutput{Body: cellToOutput(board)}, nil
}

func updateCell(ctx context.Context, queries *generated.Queries, input UpdateCellInput) (*UpdateCellOutput, error) {
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}
	cellID, err := uuidFromString(input.CellID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid cell_id", err)
	}

	cell, err := queries.UpdateCell(ctx, generated.UpdateCellParams{
		ID:      cellID,
		BoardID: boardID,
		Content: input.Body.Content,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("cell not found on this board", err)
		}
		return nil, huma.Error500InternalServerError("failed to update cell", err)
	}

	return &UpdateCellOutput{Body: cellToOutput(cell)}, nil
}

func deleteCell(ctx context.Context, queries *generated.Queries, input DeleteCellInput) (*struct{}, error) {
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}
	cellID, err := uuidFromString(input.CellID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid cell_id", err)
	}

	_, err = queries.DeleteCell(ctx, generated.DeleteCellParams{
		ID:      cellID,
		BoardID: boardID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("cell not found on this board", err)
		}
		return nil, huma.Error500InternalServerError("failed to delete cell", err)
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func cellToOutput(cell generated.Cell) CellOutput {
	return CellOutput{
		ID:      cell.ID.String(),
		BoardID: cell.BoardID.String(),
		Content: cell.Content,
	}
}
