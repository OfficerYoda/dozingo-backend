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

type UserOutput struct {
	ID       string `json:"id" format:"uuid"`
	Username string `json:"username" format:"text" maxLength:"200"`
	Email    string `json:"email" format:"text" maxLength:"200"`
}

type GetUserByIDInput struct {
	ID string `path:"id" format:"uuid"`
}

type GetUserByIDOutput struct {
	Body UserOutput `json:"user"`
}

type CreateUserInput struct {
	Body struct {
		Username string `json:"username" format:"text" maxLength:"200"`
		Email    string `json:"email" format:"text" maxLength:"200"`
	}
}

type CreateUserOutput struct {
	Body UserOutput
}

type DeleteUserInput struct {
	ID string `path:"id" format:"uuid"`
}

/// ===== Register =====

func RegisterUsers(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "get-user-by-id",
		Method:      http.MethodGet,
		Path:        "/users/{id}",
		Summary:     "Get a user by ID",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *GetUserByIDInput) (*GetUserByIDOutput, error) {
		return getUserByID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-user",
		Method:      http.MethodPost,
		Path:        "/users",
		Summary:     "Create a user",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
		return createUser(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-user",
		Method:      http.MethodDelete,
		Path:        "/users/{id}",
		Summary:     "Delete a user",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *DeleteUserInput) (*struct{}, error) {
		return deleteUser(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getUserByID(ctx context.Context, queries *generated.Queries, input GetUserByIDInput) (*GetUserByIDOutput, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id", err)
	}

	user, err := queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, huma.Error404NotFound("user not found", err)
	}
	return &GetUserByIDOutput{Body: userToOutput(user)}, nil
}

func createUser(ctx context.Context, queries *generated.Queries, input CreateUserInput) (*CreateUserOutput, error) {
	user, err := queries.CreateUser(ctx, generated.CreateUserParams{
		Username: input.Body.Username,
		Email:    input.Body.Email,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create user", err)
	}

	return &CreateUserOutput{Body: userToOutput(user)}, nil
}

func deleteUser(ctx context.Context, queries *generated.Queries, input DeleteUserInput) (*struct{}, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id", err)
	}

	_, err = queries.DeleteUser(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("user not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to delete user", err)
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func userToOutput(user generated.User) UserOutput {
	return UserOutput{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}
}
