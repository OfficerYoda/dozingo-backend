package handler

import "github.com/jackc/pgx/v5/pgtype"

func uuidFromString(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}
