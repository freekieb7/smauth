package aql

import (
	"context"
	"io"
	"net/http"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/telemetry"
)

type Service struct {
	Logger  *telemetry.Logger
	DB      *database.Database
	Builder *Builder
}

func NewService(logger *telemetry.Logger, db *database.Database, builder *Builder) Service {
	return Service{
		Logger:  logger,
		DB:      db,
		Builder: builder,
	}
}

type ColumnMetadata struct {
	Name string
	Type string
}

func (s *Service) Query(ctx context.Context, w io.Writer, query string, args map[string]any, preparedTables []PreparedTable) ([]ColumnMetadata, error) {
	if args == nil {
		args = make(map[string]any)
	}

	if preparedTables == nil {
		preparedTables = make([]PreparedTable, 0)
	}

	queryContext, err := QueryContext(query)
	if err != nil {
		s.Logger.Error("parse query error", "error", err)
		return nil, err
	}

	q, cols, err := s.Builder.BuildQuery(queryContext, args, preparedTables)
	if err != nil {
		s.Logger.Error("build query error", "error", err)
		if buildError, ok := err.(BuildError); ok {
			return nil, buildError
		}

		s.Logger.Error("internal error", "error", err)
		return nil, err
	}

	// Execute query
	rows, err := s.DB.Conn.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columnMetadata []ColumnMetadata
	for _, col := range cols {
		colName := col.Name
		if colName == "" {
			colName = "unnamed"
		}
		columnMetadata = append(columnMetadata, ColumnMetadata{
			Name: colName,
			Type: col.Type.Name(),
		})
	}

	// Stream results as JSON array
	w.Write([]byte("["))

	first := true
	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			s.Logger.Error("scan error", "error", err)
			continue
		}

		if !first {
			w.Write([]byte(","))
		}
		w.Write(jsonData)
		first = false

		// Flush each row so client receives data progressively
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	w.Write([]byte("]"))
	return columnMetadata, nil
}
