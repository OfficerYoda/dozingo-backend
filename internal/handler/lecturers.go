package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
)

/// ===== Input/Output types =====

type LecturerOutput struct {
	ID   string `json:"id" format:"uuid"`
	Name string `json:"name" required:"true" maxLength:"200"`
	Slug string `json:"slug" required:"true" maxLength:"200" pattern:"^[a-z0-9-]+$" example:"juergen-rhoetig" doc:"URL friendly identifier, lowercase letters, numbers and hyphons only"`
}

type GetLecturersOutput struct {
	Body []LecturerOutput `json:"lecturers"`
}

type GetLecturerByIDInput struct {
	ID string `path:"id" format:"uuid"`
}

type GetLecturerByIDOutput struct {
	Body LecturerOutput
}

type CreateLecturerInput struct {
	Body struct {
		Name string `json:"name" required:"true" maxLength:"200"`
		Slug string `json:"slug" required:"true" maxLength:"200" pattern:"^[a-z0-9-]+$" example:"juergen-rhoetig" doc:"URL friendly identifier, lowercase letters, numbers and hyphons only"`
	}
}

type CreateLecturerOutput struct {
	Body LecturerOutput
}

type DeleteLecturerInput struct {
	ID string `path:"id" format:"uuid"`
}

/// ===== Register =====

func RegisterLecturers(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "get-lecturers",
		Method:      http.MethodGet,
		Path:        "/lecturers",
		Summary:     "Get all lecturers",
	}, func(ctx context.Context, input *struct{}) (*GetLecturersOutput, error) {
		return getLecturers(ctx, queries)
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-lecturer",
		Method:      http.MethodGet,
		Path:        "/lecturers/{id}",
		Summary:     "Get a lecturer by ID",
	}, func(ctx context.Context, input *GetLecturerByIDInput) (*GetLecturerByIDOutput, error) {
		return getLecturer(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-lecturer",
		Method:      http.MethodPost,
		Path:        "/lecturers",
		Summary:     "Create a lecturer",
	}, func(ctx context.Context, input *CreateLecturerInput) (*CreateLecturerOutput, error) {
		return createLecturer(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-lecturer",
		Method:      http.MethodDelete,
		Path:        "/lecturers/{id}",
		Summary:     "Delete a lecturer",
	}, func(ctx context.Context, input *DeleteLecturerInput) (*struct{}, error) {
		return deleteLecturer(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getLecturers(ctx context.Context, queries *generated.Queries) (*GetLecturersOutput, error) {
	lecturers, err := queries.GetLecturers(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get lecturers")
	}

	out := &GetLecturersOutput{}
	out.Body = make([]LecturerOutput, 0)
	for _, l := range lecturers {
		out.Body = append(out.Body, lecturerToOutput(l))
	}

	return out, nil
}

func getLecturer(ctx context.Context, queries *generated.Queries, input GetLecturerByIDInput) (*GetLecturerByIDOutput, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id")
	}

	lecturer, err := queries.GetLecturersByID(ctx, id)
	if err != nil {
		return nil, huma.Error404NotFound("lecturer not found")
	}

	return &GetLecturerByIDOutput{Body: lecturerToOutput(lecturer)}, nil
}

func createLecturer(ctx context.Context, queries *generated.Queries, input CreateLecturerInput) (*CreateLecturerOutput, error) {
	lecturer, err := queries.CreateLecturer(ctx, generated.CreateLecturerParams{
		Name: input.Body.Name,
		Slug: input.Body.Slug,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create lecturer")
	}

	return &CreateLecturerOutput{Body: lecturerToOutput(lecturer)}, nil
}

func deleteLecturer(ctx context.Context, queries *generated.Queries, input DeleteLecturerInput) (*struct{}, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id")
	}

	_, err = queries.DeleteLecturer(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("lecturer not found")
		}
		return nil, huma.Error500InternalServerError("failed to delete lecturer")
	}

	return nil, nil
}

/// ===== Helper =====

func lecturerToOutput(lecturer generated.Lecturer) LecturerOutput {
	return LecturerOutput{
		ID:   lecturer.ID.String(),
		Name: lecturer.Name,
		Slug: lecturer.Slug,
	}
}
