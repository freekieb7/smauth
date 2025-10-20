package aql

import (
	"context"
	"errors"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/google/uuid"
)

type Store struct {
	DB *database.Database
}

func (s *Store) SavePreparedTable(ctx context.Context, table PreparedTable) (PreparedTable, error) {
	return table, errors.New("not implemented")
}

func (s *Store) GetPreparedTableByName(ctx context.Context, name string) (PreparedTable, error) {
	return PreparedTable{}, errors.New("not implemented")
}

func (s *Store) GetAllPreparedTables(ctx context.Context) ([]PreparedTable, error) {
	return nil, errors.New("not implemented")
}

func (s *Store) DeletePreparedTable(ctx context.Context, id uuid.UUID) error {
	return errors.New("not implemented")
}
