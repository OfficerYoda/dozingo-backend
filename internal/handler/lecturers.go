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

type LecturerOutput struct {
	ID   string `json:"id" format:"uuid"`
	Name string `json:"name" format:"text" maxLength:"200"`
	Slug string `json:"slug" format:"text" maxLength:"200" pattern:"^[a-z0-9-]+$" example:"juergen-rhoetig" doc:"URL friendly identifier, lowercase letters, numbers and hyphons only"`
}

type GetLecturersOutput struct {
	Body []LecturerOutput
}

type GetLecturerByIdentifierInput struct {
	Identifier string `path:"identifier" format:"string" doc:"can be the slug or the uuid of the lecturer"`
}

type GetLecturerByIdentifierOutput struct {
	Body LecturerOutput
}

type CreateLecturerInput struct {
	Body struct {
		Name string `json:"name" format:"text" required:"true" maxLength:"200"`
		Slug string `json:"slug" format:"text" required:"true" maxLength:"200" pattern:"^[a-z0-9-]+$" example:"juergen-rhoetig" doc:"URL friendly identifier, lowercase letters, numbers and hyphons only"`
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
		Tags:        []string{"Lecturers"},
	}, func(ctx context.Context, input *struct{}) (*GetLecturersOutput, error) {
		return getLecturers(ctx, queries)
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-lecturer-by-id",
		Method:      http.MethodGet,
		Path:        "/lecturers/{identifier}",
		Summary:     "Get a lecturer by ID or slug",
		Tags:        []string{"Lecturers"},
	}, func(ctx context.Context, input *GetLecturerByIdentifierInput) (*GetLecturerByIdentifierOutput, error) {
		return getLecturerByIdentifier(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-lecturer",
		Method:      http.MethodPost,
		Path:        "/lecturers",
		Summary:     "Create a lecturer",
		Tags:        []string{"Lecturers"},
	}, func(ctx context.Context, input *CreateLecturerInput) (*CreateLecturerOutput, error) {
		return createLecturer(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-lecturer",
		Method:      http.MethodDelete,
		Path:        "/lecturers/{id}",
		Summary:     "Delete a lecturer",
		Tags:        []string{"Lecturers"},
	}, func(ctx context.Context, input *DeleteLecturerInput) (*struct{}, error) {
		return deleteLecturer(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getLecturers(ctx context.Context, queries *generated.Queries) (*GetLecturersOutput, error) {
	lecturers, err := queries.GetLecturers(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get lecturers", err)
	}

	out := &GetLecturersOutput{}
	out.Body = make([]LecturerOutput, 0)
	for _, l := range lecturers {
		out.Body = append(out.Body, lecturerToOutput(l))
	}

	return out, nil
}

func getLecturerByIdentifier(ctx context.Context, queries *generated.Queries, input GetLecturerByIdentifierInput) (*GetLecturerByIdentifierOutput, error) {
	if uuidRegex.MatchString(input.Identifier) {
		return getLecturerByID(ctx, queries, input.Identifier)
	}
	return getLecturerBySlug(ctx, queries, input.Identifier)
}

func getLecturerByID(ctx context.Context, queries *generated.Queries, uuid string) (*GetLecturerByIdentifierOutput, error) {
	id, err := uuidFromString(uuid)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id", err)
	}

	lecturer, err := queries.GetLecturerByID(ctx, id)
	if err != nil {
		return nil, huma.Error404NotFound("lecturer not found", err)
	}
	return &GetLecturerByIdentifierOutput{Body: lecturerToOutput(lecturer)}, nil
}

func getLecturerBySlug(ctx context.Context, queries *generated.Queries, slug string) (*GetLecturerByIdentifierOutput, error) {
	lecturer, err := queries.GetLecturerBySlug(ctx, slug)
	if err != nil {
		return nil, huma.Error404NotFound("lecturer not found", err)
	}
	return &GetLecturerByIdentifierOutput{Body: lecturerToOutput(lecturer)}, nil
}

func createLecturer(ctx context.Context, queries *generated.Queries, input CreateLecturerInput) (*CreateLecturerOutput, error) {
	lecturer, err := queries.CreateLecturer(ctx, generated.CreateLecturerParams{
		Name: input.Body.Name,
		Slug: input.Body.Slug,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create lecturer", err)
	}

	return &CreateLecturerOutput{Body: lecturerToOutput(lecturer)}, nil
}

func deleteLecturer(ctx context.Context, queries *generated.Queries, input DeleteLecturerInput) (*struct{}, error) {
	id, err := uuidFromString(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id", err)
	}

	_, err = queries.DeleteLecturer(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, huma.Error404NotFound("lecturer not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to delete lecturer", err)
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func lecturerToOutput(lecturer generated.Lecturer) LecturerOutput {
	return LecturerOutput{
		ID:   lecturer.ID.String(),
		Name: lecturer.Name,
		Slug: lecturer.Slug,
	}
}
