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

type VoteOutput struct {
	ID        string `json:"id" format:"uuid"`
	UserID    string `json:"user_id" format:"uuid"`
	BoardID   string `json:"board_id" format:"uuid"`
	VoteValue string `json:"vote_value" format:"integer"`
}

type GetVotesByBoardIDInput struct {
	UserID  string `query:"user_id" format:"uuid"` // TODO eventually replace this when user auth is working
	BoardID string `path:"board_id"`
}

type GetVotesByBoardIDOutput struct {
	Body struct {
		Score     int32  `json:"score"`
		VoteCount int32  `json:"vote_count"`
		UserVote  *int32 `json:"user_vote"`
	}
}

type UpsertVoteInput struct {
	UserID  string `query:"user_id" format:"uuid" required:"true"` // TODO eventually replace this when user auth is working
	BoardID string `path:"board_id" format:"uuid"`
	Body    struct {
		VoteValue int32 `json:"vote_value" format:"integer" required:"true" minimum:"-1" maximum:"1"`
	}
}

type UpsertVoteOutput struct {
	Body VoteOutput
}

type DeleteVoteInput struct {
	UserID  string `query:"user_id" format:"uuid" required:"true"` // TODO eventually replace this when user auth is working
	BoardID string `path:"board_id" format:"uuid"`
}

/// ===== Register =====

func RegisterVotes(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	_ = queries

	huma.Register(api, huma.Operation{
		OperationID: "get-votes-by-board-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Get all votes for a Board",
		Tags:        []string{"Votes"},
	}, func(ctx context.Context, input *GetVotesByBoardIDInput) (*GetVotesByBoardIDOutput, error) {
		return getVotesByBoardID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "upsert-vote",
		Method:      http.MethodPut,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Upsert a vote",
		Description: "Update or Create a vote",
		Tags:        []string{"Votes"},
	}, func(ctx context.Context, input *UpsertVoteInput) (*UpsertVoteOutput, error) {
		return upsertVote(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-vote",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Delete a vote",
		Tags:        []string{"Votes"},
	}, func(ctx context.Context, input *DeleteVoteInput) (*struct{}, error) {
		return deleteVote(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getVotesByBoardID(ctx context.Context, queries *generated.Queries, input GetVotesByBoardIDInput) (*GetVotesByBoardIDOutput, error) {
	userID, err := uuidFromString(input.UserID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid user_id", err)
	}
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	votes, err := queries.GetVotesByBoardID(ctx, generated.GetVotesByBoardIDParams{
		UserID:  userID,
		BoardID: boardID,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error", err)
	}

	out := &GetVotesByBoardIDOutput{}
	out.Body.Score = votes.Score
	out.Body.VoteCount = votes.VoteCount
	// only return a user vote value when the user actually voted
	var userVotePtr *int32
	if votes.UserVote != 0 {
		userVotePtr = &votes.UserVote
	}
	out.Body.UserVote = userVotePtr

	return out, nil
}

func upsertVote(ctx context.Context, queries *generated.Queries, input UpsertVoteInput) (*UpsertVoteOutput, error) {
	userID, err := uuidFromString(input.UserID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid user_id", err)
	}
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	board, err := queries.UpsertVote(ctx, generated.UpsertVoteParams{
		UserID:    userID,
		BoardID:   boardID,
		VoteValue: input.Body.VoteValue,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to upsert vote", err)
	}

	return &UpsertVoteOutput{Body: voteToOutput(board)}, nil
}

func deleteVote(ctx context.Context, queries *generated.Queries, input DeleteVoteInput) (*struct{}, error) {
	userID, err := uuidFromString(input.UserID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid user_id", err)
	}
	boardID, err := uuidFromString(input.BoardID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid board_id", err)
	}

	_, err = queries.DeleteVote(ctx, generated.DeleteVoteParams{
		UserID:  userID,
		BoardID: boardID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("vote not found on this board", err)
		}
		return nil, huma.Error500InternalServerError("failed to delete vote", err)
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func voteToOutput(vote generated.Vote) VoteOutput {
	return VoteOutput{
		ID:        vote.ID.String(),
		UserID:    vote.UserID.String(),
		BoardID:   vote.BoardID.String(),
		VoteValue: string(vote.VoteValue),
	}
}
