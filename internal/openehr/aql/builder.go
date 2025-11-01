package aql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/freekieb7/smauth/internal/openehr"
	"github.com/freekieb7/smauth/internal/openehr/aql/gen"
	"github.com/freekieb7/smauth/internal/util"
)

type Builder struct{}

func NewBuilder() Builder {
	return Builder{}
}

type BuildError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (e BuildError) Error() string {
	return e.Message
}

type Parameters map[string]any

type PreparedTable struct {
	Name   string
	Source string
	Ctx    gen.ISelectQueryContext
}

type Table struct {
	Type   reflect.Type
	Name   string
	Source string
	Data   string
}

type Column struct {
	Type   reflect.Type
	Name   string
	Source string
}

var (
	EHR          = reflect.TypeFor[openehr.EHR]()
	CONTRIBUTION = reflect.TypeFor[openehr.CONTRIBUTION]()
	EHR_STATUS   = reflect.TypeFor[openehr.EHR_STATUS]()
	EHR_ACCESS   = reflect.TypeFor[openehr.EHR_ACCESS]()
	COMPOSITION  = reflect.TypeFor[openehr.COMPOSITION]()
	FOLDER       = reflect.TypeFor[openehr.FOLDER]()

	OBSERVATION = reflect.TypeFor[openehr.OBSERVATION]()
	EVALUATION  = reflect.TypeFor[openehr.EVALUATION]()
	INSTRUCTION = reflect.TypeFor[openehr.INSTRUCTION]()
	ACTION      = reflect.TypeFor[openehr.ACTION]()

	CONTENT_ITEM   = reflect.TypeFor[openehr.CONTENT_ITEM]()
	ENTRY          = reflect.TypeFor[openehr.ENTRY]()
	CARE_ENTRY     = reflect.TypeFor[openehr.CARE_ENTRY]()
	EVENT          = reflect.TypeFor[openehr.EVENT[any]]()
	ITEM_STRUCTURE = reflect.TypeFor[openehr.ITEM_STRUCTURE]()
	ITEM           = reflect.TypeFor[openehr.ITEM]()

	EVENT_CONTEXT  = reflect.TypeFor[openehr.EVENT_CONTEXT]()
	SECTION        = reflect.TypeFor[openehr.SECTION]()
	GENERIC_ENTRY  = reflect.TypeFor[openehr.GENERIC_ENTRY]()
	ADMIN_ENTRY    = reflect.TypeFor[openehr.ADMIN_ENTRY]()
	ACTIVITY       = reflect.TypeFor[openehr.ACTIVITY]()
	HISTORY        = reflect.TypeFor[openehr.HISTORY[any]]()
	POINT_EVENT    = reflect.TypeFor[openehr.POINT_EVENT[any]]()
	INTERVAL_EVENT = reflect.TypeFor[openehr.INTERVAL_EVENT[any]]()
	FEEDER_AUDIT   = reflect.TypeFor[openehr.FEEDER_AUDIT]()
	ITEM_LIST      = reflect.TypeFor[openehr.ITEM_LIST]()
	ITEM_SINGLE    = reflect.TypeFor[openehr.ITEM_SINGLE]()
	ITEM_TABLE     = reflect.TypeFor[openehr.ITEM_TABLE]()
	ITEM_TREE      = reflect.TypeFor[openehr.ITEM_TREE]()
	CLUSTER        = reflect.TypeFor[openehr.CLUSTER]()
	ELEMENT        = reflect.TypeFor[openehr.ELEMENT]()

	ROLE         = reflect.TypeFor[openehr.ROLE]()
	PERSON       = reflect.TypeFor[openehr.PERSON]()
	AGENT        = reflect.TypeFor[openehr.AGENT]()
	GROUP        = reflect.TypeFor[openehr.GROUP]()
	ORGANISATION = reflect.TypeFor[openehr.ORGANISATION]()

	PARTY_RELATIONSHIP = reflect.TypeFor[openehr.PARTY_RELATIONSHIP]()

	String  = reflect.TypeFor[string]()
	Integer = reflect.TypeFor[int]()
	Float   = reflect.TypeFor[float64]()
	Boolean = reflect.TypeFor[bool]()
	Null    = reflect.TypeFor[byte]()
)

func (b *Builder) ReflectFrom(name string) (reflect.Type, error) {
	switch name {
	case openehr.EHR_META_TYPE:
		return EHR, nil
	case openehr.CONTRIBUTION_META_TYPE:
		return CONTRIBUTION, nil
	case openehr.EHR_STATUS_META_TYPE:
		return EHR_STATUS, nil
	case openehr.EHR_ACCESS_META_TYPE:
		return EHR_ACCESS, nil
	case openehr.COMPOSITION_META_TYPE:
		return COMPOSITION, nil
	case openehr.FOLDER_META_TYPE:
		return FOLDER, nil
	case openehr.OBSERVATION_META_TYPE:
		return OBSERVATION, nil
	case openehr.EVALUATION_META_TYPE:
		return EVALUATION, nil
	case openehr.INSTRUCTION_META_TYPE:
		return INSTRUCTION, nil
	case openehr.ACTION_META_TYPE:
		return ACTION, nil
	case openehr.CONTENT_ITEM_META_TYPE:
		return CONTENT_ITEM, nil
	case openehr.ENTRY_META_TYPE:
		return ENTRY, nil
	case openehr.CARE_ENTRY_META_TYPE:
		return CARE_ENTRY, nil
	case openehr.EVENT_META_TYPE:
		return EVENT, nil
	case openehr.ITEM_STRUCTURE_META_TYPE:
		return ITEM_STRUCTURE, nil
	case openehr.ITEM_META_TYPE:
		return ITEM, nil
	case openehr.EVENT_CONTEXT_META_TYPE:
		return EVENT_CONTEXT, nil
	case openehr.SECTION_META_TYPE:
		return SECTION, nil
	case openehr.GENERIC_ENTRY_META_TYPE:
		return GENERIC_ENTRY, nil
	case openehr.ADMIN_ENTRY_META_TYPE:
		return ADMIN_ENTRY, nil
	case openehr.ACTIVITY_META_TYPE:
		return ACTIVITY, nil
	case openehr.HISTORY_META_TYPE:
		return HISTORY, nil
	case openehr.POINT_EVENT_META_TYPE:
		return POINT_EVENT, nil
	case openehr.INTERVAL_EVENT_META_TYPE:
		return INTERVAL_EVENT, nil
	case openehr.FEEDER_AUDIT_META_TYPE:
		return FEEDER_AUDIT, nil
	case openehr.ITEM_LIST_META_TYPE:
		return ITEM_LIST, nil
	case openehr.ITEM_SINGLE_META_TYPE:
		return ITEM_SINGLE, nil
	case openehr.ITEM_TABLE_META_TYPE:
		return ITEM_TABLE, nil
	case openehr.ITEM_TREE_META_TYPE:
		return ITEM_TREE, nil
	case openehr.CLUSTER_META_TYPE:
		return CLUSTER, nil
	case openehr.ELEMENT_META_TYPE:
		return ELEMENT, nil
	case openehr.ROLE_META_TYPE:
		return ROLE, nil
	case openehr.PERSON_META_TYPE:
		return PERSON, nil
	case openehr.AGENT_META_TYPE:
		return AGENT, nil
	case openehr.GROUP_META_TYPE:
		return GROUP, nil
	case openehr.ORGANISATION_META_TYPE:
		return ORGANISATION, nil
	case openehr.PARTY_RELATIONSHIP_META_TYPE:
		return PARTY_RELATIONSHIP, nil
	case "String":
		return String, nil
	case "Integer":
		return Integer, nil
	case "Real":
		return Float, nil
	case "Double":
		return Float, nil
	case "Boolean":
		return Boolean, nil
	default:
		return nil, BuildError{
			Message: "unknown type: " + name,
			Code:    "UNKNOWN_TYPE",
		}
	}
}

func (b *Builder) BuildQuery(query gen.IQueryContext, params Parameters, preparedTables []PreparedTable) (string, []Column, error) {
	switch true {
	case query.SelectQuery() != nil:
		selectQ, cols, err := b.BuildSelectQuery(query.SelectQuery(), params, preparedTables)
		if err != nil {
			return "", nil, err
		}

		colNames := ""
		for idx, col := range cols {
			if idx > 0 {
				colNames += ", "
			}
			colNames += fmt.Sprintf("to_jsonb(%s)", col.Source)
		}

		return fmt.Sprintf("SELECT array_to_json(ARRAY[%s]) FROM (%s) q", colNames, selectQ), cols, nil
	default:
		return "", nil, BuildError{
			Message: "only SELECT queries are supported",
			Code:    "NOT_IMPLEMENTED",
		}
	}
}

func (b *Builder) BuildSelectQuery(ctx gen.ISelectQueryContext, params Parameters, preparedTables []PreparedTable) (string, []Column, error) {
	// FROM clause
	fromClause, tables, err := b.BuildFromClause(ctx.FromClause(), preparedTables)
	if err != nil {
		return "", nil, err
	}

	// JOIN clauses
	extraWhereExprs := make([]string, 0)
	if ctx.AllJoinClause() != nil {
		for _, join := range ctx.AllJoinClause() {
			q, w, table, err := b.BuildJoinClause(join, params, tables)
			if err != nil {
				return "", nil, err
			}

			fromClause += " " + q
			tables = append(tables, table)

			if w != "" {
				extraWhereExprs = append(extraWhereExprs, w)
			}
		}
	}

	// WHERE clause
	var whereClause string
	if ctx.WhereClause() != nil {
		q, err := b.BuildWhereClause(ctx.WhereClause(), params, tables)
		if err != nil {
			return "", nil, err
		}
		whereClause = q
	}

	if len(extraWhereExprs) > 0 {
		if whereClause == "" {
			whereClause = "WHERE " + strings.Join(extraWhereExprs, " AND ")
		} else {
			whereClause += " AND " + strings.Join(extraWhereExprs, " AND ")
		}
	}

	// SELECT clause
	selectClause, cols, err := b.BuildSelectClause(ctx.SelectClause(), params, tables)
	if err != nil {
		return "", nil, err
	}

	// GROUP BY clause
	var groupByClause string
	if ctx.GroupByClause() != nil {
		q, err := b.BuildGroupByClause(ctx.GroupByClause(), params, tables, cols)
		if err != nil {
			return "", nil, err
		}
		groupByClause = q
	}

	// UNION clause
	var unionClause string
	if ctx.UNION() != nil {
		unionClause = "UNION "
		if ctx.ALL() != nil {
			unionClause += "ALL "
		}
		qUnion, colsUnion, err := b.BuildSelectQuery(ctx.SelectQuery(), params, preparedTables)
		if err != nil {
			return "", nil, err
		}
		if len(cols) != len(colsUnion) {
			return "", nil, BuildError{
				Message: "different number of columns in union",
				Code:    "UNION_COLUMN_MISMATCH",
			}
		}
		for i := range len(cols) {
			if cols[i].Type != colsUnion[i].Type {
				return "", nil, BuildError{
					Message: fmt.Sprintf("different column types in union: %s and %s", b.GetTypeName(cols[i].Type), b.GetTypeName(colsUnion[i].Type)),
					Code:    "UNION_COLUMN_MISMATCH",
				}
			}
		}

		unionClause += qUnion
	}

	// ORDER BY clause
	var orderByClause string
	if ctx.OrderByClause() != nil {
		q, err := b.BuildOrderByClause(ctx.OrderByClause(), tables, cols)
		if err != nil {
			return "", nil, err
		}
		orderByClause = q
	}

	// LIMIT clause
	var limitOffsetClause string
	if ctx.LimitOffsetClause() != nil {
		q, err := b.BuildLimitClause(ctx.LimitOffsetClause(), params)
		if err != nil {
			return "", nil, err
		}
		limitOffsetClause = q
	}

	return fmt.Sprintf("%s %s %s %s %s %s %s", selectClause, fromClause, whereClause, groupByClause, unionClause, orderByClause, limitOffsetClause), cols, nil
}

func (b *Builder) BuildFromClause(ctx gen.IFromClauseContext, preparedTables []PreparedTable) (string, []Table, error) {
	q, tables, err := b.BuildFromExpr(ctx.FromExpr(), preparedTables)
	if err != nil {
		return q, tables, err
	}

	return "FROM " + q, tables, nil
}

func (b *Builder) BuildFromExpr(ctx gen.IFromExprContext, preparedTables []PreparedTable) (string, []Table, error) {
	name := ctx.IDENTIFIER(0).GetText()
	t, err := b.ReflectFrom(name)
	if err != nil {
		// Check if it's an active table
		return b.BuildPreparedTable(name, preparedTables)
	}
	table := Table{
		Type:   t,
		Name:   name,
		Source: "source_0",
		Data:   "data",
	}
	if ctx.GetAlias() != nil {
		table.Name = ctx.GetAlias().GetText()
	}

	// todo maybe support ALL_VERSIONS
	// allVersions := false
	// if ctx.ALL_VERSIONS() != nil {
	// 	allVersions = true
	// }

	var q string
	switch table.Type {
	case EHR:
		q = "tbl_openehr_ehr"
	case CONTRIBUTION:
		q = "tbl_openehr_contribution"
	case COMPOSITION:
		q = "tbl_openehr_composition"
	case EHR_ACCESS:
		q = "tbl_openehr_ehr_access"
	case EHR_STATUS:
		q = "tbl_openehr_ehr_status"
	case FOLDER:
		q = "tbl_openehr_folder"
	case ROLE:
		q = "tbl_openehr_role"
	case AGENT:
		q = "tbl_openehr_agent"
	case PERSON:
		q = "tbl_openehr_person"
	case GROUP:
		q = "tbl_openehr_group"
	case ORGANISATION:
		q = "tbl_openehr_organisation"
	default:
		return "", nil, BuildError{
			Message: "cannot use type in FROM clause: " + table.Type.Name(),
			Code:    "FROM_CLAUSE_TYPE_MISMATCH",
		}
	}
	q += " " + table.Source

	return q, []Table{table}, nil
}

func (b *Builder) BuildPreparedTable(name string, preparedTables []PreparedTable) (string, []Table, error) {
	for _, pt := range preparedTables {
		if pt.Name == name {
			_, cols, err := b.BuildSelectQuery(pt.Ctx, Parameters{}, make([]PreparedTable, 0)) // Validate the prepared table's AQL
			if err != nil {
				return "", nil, err
			}

			tables := make([]Table, len(cols))
			for i, col := range cols {
				tables[i] = Table{
					Type:   col.Type,
					Name:   col.Name,
					Source: pt.Source,
					Data:   col.Source,
				}
			}

			return pt.Source, tables, nil
		}
	}
	return "", nil, BuildError{
		Message: "unknown active table: " + name,
		Code:    "UNKNOWN_ACTIVE_TABLE",
	}
}

func (b *Builder) BuildJoinClause(ctx gen.IJoinClauseContext, params Parameters, tables []Table) (string, string, Table, error) {
	q := "JOIN "
	if ctx.LEFT() != nil {
		q = "LEFT JOIN "
	}

	qExpr, wExpr, table, err := b.BuildJoinExpr(ctx.JoinExpr(), params, tables)
	if err != nil {
		return "", "", table, err
	}

	q += " " + qExpr
	return q, wExpr, table, nil
}

func (b *Builder) BuildJoinExpr(ctx gen.IJoinExprContext, params Parameters, tables []Table) (string, string, Table, error) {
	name := ctx.IDENTIFIER(0).GetText()
	t, err := b.ReflectFrom(name)
	if err != nil {
		return "", "", Table{}, err
	}

	if t == String || t == Integer || t == Float || t == Boolean || t == Null {
		return "", "", Table{}, BuildError{
			Message: fmt.Sprintf("cannot use %s in JOIN clause", t.Name()),
			Code:    "INVALID_TYPE",
		}
	}

	targetTable := Table{
		Type:   t,
		Name:   name,
		Source: fmt.Sprintf("source_%d", len(tables)),
		Data:   "data",
	}
	if ctx.GetAlias() != nil {
		targetTable.Name = ctx.GetAlias().GetText()
	}

	switch true {
	case ctx.ON() != nil:
		{
			var sourceTable Table
			sourceName := ctx.GetSource().GetText()
			found := false
			for _, table := range tables {
				if sourceName == table.Name {
					sourceTable = table
					found = true
					break
				}
			}
			if !found {
				return "", "", targetTable, BuildError{
					Message: "unknown table in JOIN: " + sourceName,
					Code:    "UNKNOWN_TABLE",
				}
			}

			allVersions := ctx.ALL_VERSIONS() != nil

			var q string
			switch sourceTable.Type {
			case EHR:
				switch targetTable.Type {
				case CONTRIBUTION:
					q = fmt.Sprintf("tbl_contribution %s ON %s.id = %s.ehr_id", targetTable.Source, sourceTable.Source, targetTable.Source)
				case COMPOSITION:
					q = "tbl_composition_current"
					if allVersions {
						q = "tbl_composition"
					}
					q = fmt.Sprintf("%s %s ON %s.id = %s.ehr_id", q, targetTable.Source, sourceTable.Source, targetTable.Source)
				case EHR_ACCESS:
					q = "tbl_ehr_access_current"
					if allVersions {
						q = "tbl_ehr_access"
					}
					q = fmt.Sprintf("%s %s ON %s.id = %s.ehr_id", q, targetTable.Source, sourceTable.Source, targetTable.Source)
				case EHR_STATUS:
					q = "tbl_ehr_status_current"
					if allVersions {
						q = "tbl_ehr_status"
					}
					q = fmt.Sprintf("%s %s ON %s.id = %s.ehr_id", q, targetTable.Source, sourceTable.Source, targetTable.Source)
				case FOLDER:
					q = "tbl_folder_current"
					if allVersions {
						q = "tbl_folder"
					}
					q = fmt.Sprintf("%s %s ON %s.id = %s.ehr_id", q, targetTable.Source, sourceTable.Source, targetTable.Source)
				case PERSON:
					q = "tbl_person_current"
					if allVersions {
						q = "tbl_person"
					}
					q = fmt.Sprintf("%s %s ON %s.party_id = %s.id", q, targetTable.Source, sourceTable.Source, targetTable.Source)
				default:
					return "", "", targetTable, BuildError{
						Message: "cannot join EHR with " + targetTable.Type.Name(),
						Code:    "UNKNOWN_TABLE",
					}
				}
			case PERSON:
				switch targetTable.Type {
				case EHR:
					q = fmt.Sprintf("tbl_ehr %s ON %s.id = %s.party_id", targetTable.Source, sourceTable.Source, targetTable.Source)
				case GROUP:
					q = "tbl_group_current"
					if allVersions {
						q = "tbl_group"
					}
					q = fmt.Sprintf("tbl_party_relationship tmp_%s ON %s.id = tmp_%s.source_id JOIN %s %s ON tmp_%s.target_id = %s.id", targetTable.Source, sourceTable.Source, targetTable.Source, q, targetTable.Source, targetTable.Source, targetTable.Source)
				case PERSON:
					q = "tbl_group_current"
					if allVersions {
						q = "tbl_group"
					}
					q = fmt.Sprintf("tbl_party_relationship tmp_%s ON %s.id = tmp_%s.source_id JOIN %s %s ON tmp_%s.target_id = %s.id", targetTable.Source, sourceTable.Source, targetTable.Source, q, targetTable.Source, targetTable.Source, targetTable.Source)
				default:
					return "", "", targetTable, BuildError{
						Message: "cannot join EHR with " + targetTable.Type.Name(),
						Code:    "UNKNOWN_TABLE",
					}
				}
			case PARTY_RELATIONSHIP:
				switch targetTable.Type {
				case GROUP:
					q = "tbl_group_current"
					if allVersions {
						q = "tbl_group"
					}
					q = fmt.Sprintf("%s %s ON (%s.%s -> 'target' -> 'id' ->> 'value') = %s.id", q, targetTable.Source, sourceTable.Source, sourceTable.Data, targetTable.Source)
				}
			default:
				return "", "", targetTable, BuildError{
					Message: "cannot join from type: " + sourceTable.Type.Name(),
					Code:    "UNKNOWN_TABLE",
				}
			}

			return q, "", targetTable, nil

		}
	case ctx.IN() != nil:
		{
			var sourceTable Table
			sourceName := ctx.GetSource().GetText()
			found := false
			for _, table := range tables {
				if sourceName == table.Name {
					sourceTable = table
					found = true
					break
				}
			}
			if !found {
				return "", "", targetTable, BuildError{
					Message: "unknown table in IN clause: " + sourceName,
					Code:    "UNKNOWN_TABLE",
				}
			}

			q := fmt.Sprintf("LATERAL (SELECT * FROM JSON_TABLE(%s.%s, 'strict $.*.** ? (@._type == \"%s\")' COLUMNS(%s JSONB PATH '$'))) %s ON TRUE", sourceTable.Source, sourceTable.Data, targetTable.Type.Name(), targetTable.Data, targetTable.Source)
			w := fmt.Sprintf("jsonb_path_query_array(%s.%s, '$.*.**._type') @> '[\"%s\"]'::jsonb", sourceTable.Source, sourceTable.Data, targetTable.Type.Name())
			return q, w, targetTable, nil
		}
	case ctx.AT() != nil:
		{
			sourceTable, path, typ, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", "", targetTable, err
			}

			// Only allow joining on complex types
			if typ == String || typ == Integer || typ == Float || typ == Boolean || typ == Null {
				return "", "", targetTable, BuildError{
					Message: fmt.Sprintf("cannot use %s in AT clause", typ.Name()),
					Code:    "INVALID_TYPE",
				}
			}

			q := fmt.Sprintf("LATERAL (SELECT * FROM JSON_TABLE(%s.%s, '$%s ? (@._type == \"%s\")' COLUMNS(%s JSONB PATH '$'))) %s ON TRUE", sourceTable.Source, sourceTable.Data, path, targetTable.Type.Name(), targetTable.Data, targetTable.Source)
			return q, "", targetTable, nil
		}
	default:
		{
			return "", "", Table{}, errors.New("unknown join expression")
		}
	}
}

func (b *Builder) BuildWhereClause(ctx gen.IWhereClauseContext, params Parameters, tables []Table) (string, error) {
	q, err := b.BuildWhereExpr(ctx.WhereExpr(), params, tables)
	if err != nil {
		return "", err
	}

	return "WHERE " + q, nil
}

func (b *Builder) BuildWhereExpr(ctx gen.IWhereExprContext, params Parameters, tables []Table) (string, error) {
	switch true {
	case ctx.BooleanCondition() != nil:
		{
			return b.BuildBooleanCondition(ctx.BooleanCondition(), params, tables)
		}
	case ctx.NOT() != nil:
		{
			q, err := b.BuildWhereExpr(ctx.WhereExpr(0), params, tables)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("NOT (%s)", q), nil
		}
	case ctx.AND() != nil:
		{
			qLeft, err := b.BuildWhereExpr(ctx.WhereExpr(0), params, tables)
			if err != nil {
				return "", err
			}
			qRight, err := b.BuildWhereExpr(ctx.WhereExpr(1), params, tables)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("(%s) AND (%s)", qLeft, qRight), nil
		}
	case ctx.OR() != nil:
		{
			qLeft, err := b.BuildWhereExpr(ctx.WhereExpr(0), params, tables)
			if err != nil {
				return "", err
			}
			qRight, err := b.BuildWhereExpr(ctx.WhereExpr(1), params, tables)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("(%s) OR (%s)", qLeft, qRight), nil
		}
	case ctx.SYM_LEFT_PAREN() != nil:
		{
			return b.BuildWhereExpr(ctx.WhereExpr(0), params, tables)
		}
	default:
		{
			return "", errors.New("unknown where expression")
		}
	}
}

func (b *Builder) BuildGroupByClause(ctx gen.IGroupByClauseContext, params Parameters, tables []Table, cols []Column) (string, error) {
	q := "GROUP BY "

out:
	for i, path := range ctx.AllIdentifiedPath() {
		if i > 0 {
			q += ", "
		}

		// Find in columns first
		pathText := path.GetText()
		for _, col := range cols {
			if col.Name == pathText {
				q += col.Source
				continue out
			}
		}

		// Not found in columns, build path
		table, path, _, err := b.BuildIdentifiedPath(path, params, tables)
		if err != nil {
			return "", err
		}

		if path == "" {
			q += table.Source + "." + table.Data
			continue
		}

		q += fmt.Sprintf("jsonb_path_query_first(%s.%s, '$%s')", table.Source, table.Data, path)
	}
	return q, nil
}

func (b *Builder) BuildBooleanCondition(ctx gen.IBooleanConditionContext, params Parameters, tables []Table) (string, error) {
	switch true {
	case ctx.EXISTS() != nil:
		{
			source, path, _, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("jsonb_path_exists(%s.%s, '$%s')", source.Source, source.Data, path), nil
		}
	case ctx.COMPARISON_OPERATOR() != nil:
		{
			qLeft, typLeft, err := b.BuildComparisonOperand(ctx.ComparisonOperand(0), params, tables)
			if err != nil {
				return "", err
			}
			qRight, typRight, err := b.BuildComparisonOperand(ctx.ComparisonOperand(1), params, tables)
			if err != nil {
				return "", err
			}
			if typLeft != typRight {
				return "", BuildError{
					Message: fmt.Sprintf("cannot compare different data types: %s and %s", typLeft.Name(), typRight.Name()),
					Code:    "TYPE_MISMATCH",
				}
			}
			return fmt.Sprintf("%s %s %s", qLeft, ctx.COMPARISON_OPERATOR().GetText(), qRight), nil
		}
	case ctx.LIKE() != nil:
		{
			qLeft, typLeft, err := b.BuildComparisonOperand(ctx.ComparisonOperand(0), params, tables)
			if err != nil {
				return "", err
			}
			if typLeft != String {
				return "", BuildError{
					Message: fmt.Sprintf("cannot compare different data types: %s and %s", typLeft.Name(), String.Name()),
					Code:    "TYPE_MISMATCH",
				}
			}

			qRight, err := b.BuildStringOperand(ctx.StringOperand(), params, tables)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("%s LIKE '%s'", qLeft, qRight), nil
		}
	case ctx.IN() != nil:
		{
			qLeft, typLeft, err := b.BuildComparisonOperand(ctx.ComparisonOperand(0), params, tables)
			if err != nil {
				return "", err
			}

			qRight, typRight, err := b.BuildInOperand(ctx.InOperand(), params)
			if err != nil {
				return "", err
			}
			if typLeft != typRight {
				return "", BuildError{
					Message: fmt.Sprintf("cannot compare different data types: %s and %s", typLeft.Name(), typRight.Name()),
					Code:    "TYPE_MISMATCH",
				}
			}

			return fmt.Sprintf("%s IN (%s)", qLeft, qRight), nil

		}
	case ctx.CONTAINS() != nil:
		{
			sourceName := ctx.IDENTIFIER(0).GetText()
			targetName := ctx.IDENTIFIER(1).GetText()

			for _, table := range tables {
				if table.Name == sourceName {
					return fmt.Sprintf("%s.%s @? '$.*.** ? (@._type == \"%s\")'", table.Source, table.Data, targetName), nil
				}
			}

			return "", BuildError{
				Message: fmt.Sprintf("unknown table in CONTAINS clause: %s", sourceName),
				Code:    "UNKNOWN_TABLE",
			}

		}
	case ctx.SYM_LEFT_PAREN() != nil:
		{
			return b.BuildBooleanCondition(ctx.BooleanCondition(), params, tables)
		}
	default:
		{
			return "", errors.New("unknown boolean condition")
		}
	}
}

func (b *Builder) BuildInOperand(ctx gen.IInOperandContext, params Parameters) (string, reflect.Type, error) {
	switch true {
	case ctx.SelectQuery() != nil:
		{
			q, cols, err := b.BuildSelectQuery(ctx.SelectQuery(), params, make([]PreparedTable, 0))
			if err != nil {
				return "", nil, err
			}
			if len(cols) != 1 {
				return "", nil, BuildError{
					Message: "subquery in IN must return exactly one column",
					Code:    "SUBQUERY_COLUMN_MISMATCH",
				}
			}
			return fmt.Sprintf("(%s)", q), cols[0].Type, nil
		}
	case ctx.AllInOperandValue() != nil:
		{
			q := ""
			var typ reflect.Type
			for i, inOp := range ctx.AllInOperandValue() {
				if i > 0 {
					q += ", "
				}
				qIn, tIn, err := b.BuildInOperandValue(inOp, params)
				if err != nil {
					return "", nil, err
				}
				if i == 0 {
					typ = tIn
				}
				if tIn != typ {
					return "", nil, BuildError{
						Message: "all IN operands must be of the same type",
						Code:    "TYPE_MISMATCH",
					}
				}
				q += qIn
			}
			return q, typ, nil
		}
	default:
		{
			return "", nil, errors.New("unknown IN operand")
		}
	}
}

func (b *Builder) BuildInOperandValue(ctx gen.IInOperandValueContext, params Parameters) (string, reflect.Type, error) {
	switch true {
	case ctx.Primitive() != nil:
		{
			q, typ, err := b.BuildPrimitive(ctx.Primitive())
			if err != nil {
				return "", nil, err
			}
			if typ == String {
				q = fmt.Sprintf("'\"%s\"'::jsonb", strings.ReplaceAll(q, `"`, `\"`))
			} else {
				q = fmt.Sprintf("to_jsonb(%s)", q)
			}
			return q, typ, nil
		}
	case ctx.PARAMETER() != nil:
		{
			return b.BuildParameter(ctx.PARAMETER(), params)
		}
	default:
		{
			return "", nil, errors.New("unknown IN operand")
		}
	}
}

func (b *Builder) BuildStringOperand(ctx gen.IStringOperandContext, params Parameters, tables []Table) (string, error) {
	switch true {
	case ctx.STRING() != nil:
		{
			return ctx.STRING().GetText(), nil
		}
	case ctx.PARAMETER() != nil:
		{
			q, t, err := b.BuildParameter(ctx.PARAMETER(), params)
			if err != nil {
				return "", err
			}

			if t != String {
				return "", BuildError{
					Message: fmt.Sprintf("cannot use parameter of type %s as string", t.Name()),
					Code:    "TYPE_MISMATCH",
				}
			}

			return q, nil
		}
	case ctx.IDENTIFIER() != nil:
		{
			name := ctx.IDENTIFIER().GetText()
			for _, table := range tables {
				if name == table.Name {
					if table.Type != String {
						return "", errors.New("wrong data type")
					}
					return fmt.Sprintf("%s.%s", table.Source, table.Data), nil
				}
			}

			return "", BuildError{
				Message: fmt.Sprintf("unknown identifier in string operand: %s", name),
				Code:    "UNKNOWN_IDENTIFIER",
			}
		}
	default:
		{
			return "", errors.New("unknown string operand")
		}
	}
}

func (b *Builder) BuildComparisonOperand(ctx gen.IComparisonOperandContext, params Parameters, tables []Table) (string, reflect.Type, error) {
	switch true {
	case ctx.Primitive() != nil:
		{
			q, t, err := b.BuildPrimitive(ctx.Primitive())
			if err != nil {
				return "", nil, err
			}
			if t == String {
				q = fmt.Sprintf("'\"%s\"'::jsonb", strings.ReplaceAll(q, `"`, `\"`))
			} else {
				q = fmt.Sprintf("to_jsonb(%s)", q)
			}
			return q, t, nil
		}
	case ctx.PARAMETER() != nil:
		{
			q, t, err := b.BuildParameter(ctx.PARAMETER(), params)
			if err != nil {
				return "", t, err
			}

			if t == String {
				q = fmt.Sprintf("'\"%s\"'::jsonb", strings.ReplaceAll(q, `"`, `\"`))
			} else {
				q = fmt.Sprintf("to_jsonb(%s)", q)
			}

			return q, t, nil
		}
	case ctx.IdentifiedPath() != nil:
		{
			table, path, t, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", t, err
			}

			if t == nil {
				return "", t, BuildError{
					Message: "cannot use complex type in comparison",
					Code:    "TYPE_MISMATCH",
				}
			}

			q := fmt.Sprintf("jsonb_path_query_first(%s.%s, '$%s')", table.Source, table.Data, path)
			return q, t, nil
		}
	default:
		{
			return "", nil, errors.New("unknown comparison operand")
		}
	}
}

func (b *Builder) BuildSelectClause(ctx gen.ISelectClauseContext, params Parameters, tables []Table) (string, []Column, error) {
	q := "SELECT "
	if ctx.DISTINCT() != nil {
		q += "DISTINCT "
	}

	cols := make([]Column, 0)
	for i, selectExpr := range ctx.AllSelectExpr() {
		if i > 0 {
			q += ", "
		}

		exprQ, moreCols, err := b.BuildSelectExpr(selectExpr, params, tables, len(cols))
		if err != nil {
			return q, cols, err
		}

		q += exprQ
		cols = append(cols, moreCols...)
	}

	return q, cols, nil
}

func (b *Builder) BuildSelectExpr(ctx gen.ISelectExprContext, params Parameters, tables []Table, colNr int) (string, []Column, error) {
	switch true {
	case ctx.SYM_ASTERISK() != nil:
		{
			q := ""
			cols := make([]Column, len(tables))
			for i, table := range tables {
				if i > 0 {
					q += ","
				}

				cols[i] = Column{
					Type:   table.Type,
					Name:   table.Name,
					Source: fmt.Sprintf("col_%d", colNr),
				}

				colNr++
				q += fmt.Sprintf("%s.%s %s", table.Source, table.Data, cols[i].Source)
			}

			return q, cols, nil
		}
	case ctx.ColumnExpr() != nil:
		{
			q, t, err := b.BuildColumnExpr(ctx.ColumnExpr(), params, tables)
			if err != nil {
				return "", nil, err
			}

			if t == nil {
				return "", nil, errors.New("cannot select complex type")
			}

			name := ""
			if ctx.IDENTIFIER() != nil {
				name = ctx.IDENTIFIER().GetText()
			}

			col := Column{
				Type:   t,
				Name:   name,
				Source: fmt.Sprintf("col_%d", colNr),
			}

			return fmt.Sprintf("%s %s", q, col.Source), []Column{col}, nil
		}
	default:
		{
			return "", nil, errors.New("unknown select expression")
		}
	}
}

func (b *Builder) BuildColumnExpr(ctx gen.IColumnExprContext, params Parameters, tables []Table) (string, reflect.Type, error) {
	switch true {
	case ctx.Primitive() != nil:
		{
			q, t, err := b.BuildPrimitive(ctx.Primitive())
			if err != nil {
				return "", nil, err
			}

			if t == String {
				q = fmt.Sprintf("'\"%s\"'::jsonb", strings.ReplaceAll(q, `"`, `\"`))
			} else {
				q = fmt.Sprintf("to_jsonb(%s)", q)
			}

			return q, t, nil
		}
	case ctx.PARAMETER() != nil:
		{
			q, t, err := b.BuildParameter(ctx.PARAMETER(), params)
			if err != nil {
				return "", nil, err
			}

			if t == String {
				q = fmt.Sprintf("'\"%s\"'::jsonb", strings.ReplaceAll(q, `"`, `\"`))
			} else {
				q = fmt.Sprintf("to_jsonb(%s)", q)
			}

			return q, t, nil
		}
	case ctx.IdentifiedPath() != nil:
		{
			table, path, t, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", nil, err
			}

			if t == nil {
				return "", t, BuildError{
					Message: "cannot use complex type in select",
					Code:    "TYPE_MISMATCH",
				}
			}

			// Unfold path
			if path != "" {
				return fmt.Sprintf("jsonb_path_query(%s.%s, '$%s')", table.Source, table.Data, path), t, nil
			}

			return fmt.Sprintf("%s.%s", table.Source, table.Data), t, nil
		}
	case ctx.AggregateFunctionCall() != nil:
		{
			q, typ, err := b.BuildAggregateFunctionCall(ctx.AggregateFunctionCall(), params, tables)
			if err != nil {
				return "", nil, err
			}
			return q, typ, nil
		}
	case ctx.FunctionCall() != nil:
		{
			// todo
			return "", nil, errors.New("function call not implemented")
		}
	default:
		{
			return "", nil, errors.New("unknown column expression")
		}
	}
}

func (b *Builder) BuildOrderByClause(ctx gen.IOrderByClauseContext, tables []Table, cols []Column) (string, error) {
	q := "ORDER BY "
	for idx, expr := range ctx.AllOrderByExpr() {
		if idx > 0 {
			q += ","
		}

		qExpr, err := b.BuildOrderByExpr(expr, tables, cols)
		if err != nil {
			return "", err
		}
		q += qExpr
	}
	return q, nil
}

func (b *Builder) BuildOrderByExpr(ctx gen.IOrderByExprContext, tables []Table, cols []Column) (string, error) {
	name := ctx.IDENTIFIER().GetText()
	order := "ASC"
	if ctx.DESC() != nil {
		order = "DESC"
	}

	// First try cols
	for _, col := range cols {
		if col.Name == name {
			return fmt.Sprintf("\"%s\" %s", col.Name, order), nil
		}
	}

	// Next try tables
	for _, table := range tables {
		if table.Name == name {
			return fmt.Sprintf("%s.%s %s", table.Source, table.Data, order), nil
		}
	}

	return "", BuildError{
		Message: fmt.Sprintf("unknown identifier in ORDER BY clause: %s", name),
		Code:    "UNKNOWN_IDENTIFIER",
	}
}

func (b *Builder) BuildLimitClause(ctx gen.ILimitOffsetClauseContext, params Parameters) (string, error) {
	q := ""
	if ctx.GetLeftLimit() != nil {
		limit, err := b.BuildLimitOperand(ctx.LimitOperand(0), params)
		if err != nil {
			return "", nil
		}
		q += "LIMIT " + limit
	}
	if ctx.GetRightLimit() != nil {
		limit, err := b.BuildLimitOperand(ctx.LimitOperand(1), params)
		if err != nil {
			return "", nil
		}
		q += "LIMIT " + limit
	}
	if ctx.GetLeftOffset() != nil {
		offset, err := b.BuildLimitOperand(ctx.LimitOperand(0), params)
		if err != nil {
			return "", nil
		}
		q += "OFFSET " + offset
	}
	if ctx.GetRightOffset() != nil {
		offset, err := b.BuildLimitOperand(ctx.LimitOperand(1), params)
		if err != nil {
			return "", nil
		}
		q += "OFFSET " + offset
	}
	return q, nil
}

func (b *Builder) BuildLimitOperand(ctx gen.ILimitOperandContext, params Parameters) (string, error) {
	switch true {
	case ctx.INTEGER() != nil:
		{
			return ctx.INTEGER().GetText(), nil
		}
	case ctx.PARAMETER() != nil:
		{
			q, t, err := b.BuildParameter(ctx.PARAMETER(), params)
			if err != nil {
				return "", err
			}
			if t != Integer {
				return "", BuildError{
					Message: fmt.Sprintf("cannot use parameter of type %s as integer", t.Name()),
					Code:    "TYPE_MISMATCH",
				}
			}
			return q, nil
		}
	default:
		{
			return "", errors.New("unknown limit operand")
		}
	}
}

func (b *Builder) BuildAggregateFunctionCall(ctx gen.IAggregateFunctionCallContext, params Parameters, tables []Table) (string, reflect.Type, error) {
	switch true {
	case ctx.COUNT() != nil:
		{
			switch true {
			case ctx.SYM_ASTERISK() != nil:
				if ctx.DISTINCT() != nil {
					return "COUNT(DISTINCT *)", Integer, nil
				}
				return "COUNT(*)", Integer, nil
			case ctx.IdentifiedPath() != nil:
				table, path, _, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
				if err != nil {
					return "", nil, err
				}

				if ctx.DISTINCT() != nil {
					return fmt.Sprintf("COUNT(DISTINCT jsonb_path_query_first(%s.%s, '$%s'))", table.Source, table.Data, path), Integer, nil
				}
				return fmt.Sprintf("COUNT(jsonb_path_query_first(%s.%s, '$%s'))", table.Source, table.Data, path), Integer, nil
			default:
				return "", nil, errors.New("unknown COUNT argument")
			}
		}
	case ctx.SUM() != nil:
		{
			table, path, typ, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", nil, err
			}

			if typ != Integer && typ != Float {
				return "", nil, BuildError{
					Message: "cannot use SUM on type " + typ.Name(),
					Code:    "TYPE_MISMATCH",
				}
			}

			return fmt.Sprintf("SUM(jsonb_path_query_first(%s.%s, '$%s'))", table.Source, table.Data, path), typ, nil
		}
	case ctx.AVG() != nil:
		{
			table, path, typ, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", nil, err
			}

			if typ != Integer && typ != Float {
				return "", nil, BuildError{
					Message: "cannot use AVG on type " + typ.Name(),
					Code:    "TYPE_MISMATCH",
				}
			}

			return fmt.Sprintf("AVG(jsonb_path_query_first(%s.%s, '$%s'))", table.Source, table.Data, path), typ, nil
		}
	case ctx.MIN() != nil:
		{
			table, path, typ, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", nil, err
			}

			if typ != Integer && typ != Float {
				return "", nil, BuildError{
					Message: "cannot use MIN on type " + typ.Name(),
					Code:    "TYPE_MISMATCH",
				}
			}

			return fmt.Sprintf("MIN(jsonb_path_query_first(%s.%s, '$%s'))", table.Source, table.Data, path), typ, nil
		}
	case ctx.MAX() != nil:
		{
			table, path, typ, err := b.BuildIdentifiedPath(ctx.IdentifiedPath(), params, tables)
			if err != nil {
				return "", nil, err
			}

			if typ != Integer && typ != Float {
				return "", nil, BuildError{
					Message: "cannot use MAX on type " + typ.Name(),
					Code:    "TYPE_MISMATCH",
				}
			}

			return fmt.Sprintf("MAX(jsonb_path_query_first(%s.%s, '$%s'))", table.Source, table.Data, path), typ, nil
		}
	default:
		{
			return "", nil, errors.New("unknown aggregate function")
		}
	}
}

func (b *Builder) BuildPrimitive(ctx gen.IPrimitiveContext) (string, reflect.Type, error) {
	switch true {
	case ctx.STRING() != nil:
		{
			s := ctx.GetText()
			s = s[1 : len(s)-1] // Strip quotes
			return s, String, nil
		}
	case ctx.IntPrimitive() != nil:
		{
			return ctx.GetText(), Integer, nil
		}
	case ctx.FloatPrimitive() != nil:
		{
			return ctx.GetText(), Float, nil
		}
	case ctx.BOOLEAN() != nil:
		{
			return strings.ToLower(ctx.GetText()), Boolean, nil
		}
	case ctx.NULL() != nil:
		{
			return "null", Null, nil
		}
	default:
		{
			return "", nil, errors.New("unknown primitive type")
		}
	}
}

func (b *Builder) BuildParameter(ctx antlr.TerminalNode, params Parameters) (string, reflect.Type, error) {
	name := ctx.GetText()[1:] // Strip $

	v, found := params[name]
	if !found {
		return "", nil, BuildError{
			Message: "unknown parameter: " + name,
			Code:    "UNKNOWN_PARAMETER",
		}
	}

	switch t := v.(type) {
	case string:
		{
			return t, String, nil
		}
	case int8, int16, int32, int64:
		{
			return fmt.Sprintf("%d", t), Integer, nil
		}
	case float32, float64:
		{
			return fmt.Sprintf("%d", t), Float, nil
		}
	case bool:
		{
			s := "false"
			if t {
				s = "true"
			}
			return s, Boolean, nil
		}
	case nil:
		{
			return "null", Null, nil
		}
	default:
		{
			return "", nil, BuildError{
				Message: fmt.Sprintf("unsupported parameter type: %T", v),
				Code:    "TYPE_MISMATCH",
			}
		}
	}
}

func (b *Builder) BuildIdentifiedPath(ctx gen.IIdentifiedPathContext, params Parameters, tables []Table) (Table, string, reflect.Type, error) {
	q := ""
	name := ctx.IDENTIFIER().GetText()

	var sourceTable Table
	found := false
	for _, table := range tables {
		if name == table.Name {
			sourceTable = table
			found = true
			break
		}
	}
	if !found {
		return sourceTable, "", nil, BuildError{
			Message: "unknown table in identified path: " + name,
			Code:    "UNKNOWN_TABLE",
		}
	}
	t := sourceTable.Type

	if ctx.PathCondition() != nil {
		qCond, err := b.BuildPathCondition(ctx.PathCondition(), params, t)
		if err != nil {
			return sourceTable, "", nil, err
		}
		q += fmt.Sprintf(" ? (%s)", qCond)
	}

	if ctx.ObjectPath() != nil {
		qPath, typPath, err := b.BuildObjectPath(ctx.ObjectPath(), params, t)
		if err != nil {
			return sourceTable, "", nil, err
		}

		q += qPath
		t = typPath
	}

	if ctx.CAST() != nil {
		castType, err := b.ReflectFrom(ctx.CAST().GetText()[2:])
		if err != nil {
			return sourceTable, "", nil, err
		}

		if t != castType {
			return sourceTable, "", nil, BuildError{
				Message: fmt.Sprintf("cannot cast from %s to %s", t.Name(), castType.Name()),
				Code:    "TYPE_MISMATCH",
			}
		}
	}

	return sourceTable, q, t, nil
}

func (b *Builder) BuildObjectPath(ctx gen.IObjectPathContext, params Parameters, t reflect.Type) (string, reflect.Type, error) {
	q := ""
	for _, part := range ctx.AllPathPart() {
		qPart, typPart, err := b.BuildPathPart(part, params, t)
		if err != nil {
			return "", t, err
		}
		q += qPart
		t = typPart
	}
	return q, t, nil
}

func (b *Builder) BuildPathPart(ctx gen.IPathPartContext, params Parameters, t reflect.Type) (string, reflect.Type, error) {
	path := ctx.IDENTIFIER().GetText()

	pathType, found := b.GetFieldTypeByJSONTag(t, path)
	if !found {
		return "", nil, BuildError{
			Message: fmt.Sprintf("unknown path part '%s' in type %s", path, b.GetTypeName(t)),
			Code:    "UNKNOWN_PATH",
		}
	}

	q := "." + path
	if ctx.PathCondition() != nil {
		qCond, err := b.BuildPathCondition(ctx.PathCondition(), params, pathType)
		if err != nil {
			return "", nil, err
		}
		q += fmt.Sprintf(" ? (@%s)", qCond)
	}
	return q, pathType, nil
}

func (b *Builder) BuildPathCondition(ctx gen.IPathConditionContext, params Parameters, t reflect.Type) (string, error) {
	switch true {
	case ctx.PathConditionOperand(0) != nil:
		{
			qLeft, typLeft, err := b.BuildPathConditionOperand(ctx.PathConditionOperand(0), params, t)
			if err != nil {
				return "", err
			}
			qRight, typRight, err := b.BuildPathConditionOperand(ctx.PathConditionOperand(1), params, t)
			if err != nil {
				return "", err
			}

			if typLeft != typRight {
				return "", BuildError{
					Message: fmt.Sprintf("cannot compare different data types: %s and %s", typLeft.Name(), typRight.Name()),
					Code:    "TYPE_MISMATCH",
				}
			}

			comp := ctx.COMPARISON_OPERATOR().GetText()
			if comp == "=" {
				comp = "==" // JSONPath uses '==' for equality
			}

			return fmt.Sprintf("@%s %s @%s", qLeft, comp, qRight), nil
		}
	case ctx.AND() != nil:
		{
			left, err := b.BuildPathCondition(ctx.PathCondition(0), params, t)
			if err != nil {
				return "", err
			}
			right, err := b.BuildPathCondition(ctx.PathCondition(1), params, t)
			if err != nil {
				return "", err
			}
			return left + " && " + right, nil
		}
	case ctx.OR() != nil:
		{
			left, err := b.BuildPathCondition(ctx.PathCondition(0), params, t)
			if err != nil {
				return "", err
			}
			right, err := b.BuildPathCondition(ctx.PathCondition(1), params, t)
			if err != nil {
				return "", err
			}
			return left + " || " + right, nil
		}
	case ctx.SYM_LEFT_PAREN() != nil:
		{
			return b.BuildPathCondition(ctx.PathCondition(0), params, t)
		}
	default:
		{
			return "", errors.New("unknown path condition")
		}
	}
}

func (b *Builder) BuildPathConditionOperand(ctx gen.IPathConditionOperandContext, params Parameters, t reflect.Type) (string, reflect.Type, error) {
	switch true {
	case ctx.Primitive() != nil:
		{
			q, typ, err := b.BuildPrimitive(ctx.Primitive())
			if err != nil {
				return "", nil, err
			}
			if typ == String {
				q = fmt.Sprintf("\"%s\"", q)
			}
			return q, typ, nil
		}
	case ctx.ObjectPath() != nil:
		{
			return b.BuildObjectPath(ctx.ObjectPath(), params, t)
		}
	case ctx.PARAMETER() != nil:
		{
			q, typ, err := b.BuildParameter(ctx.PARAMETER(), params)
			if err != nil {
				return "", nil, err
			}
			if typ == String {
				q = fmt.Sprintf("\"%s\"", q)
			}
			return q, typ, nil
		}
	default:
		{
			return "", nil, errors.New("unknown path condition operand")
		}
	}
}

func (b *Builder) GetTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	val := reflect.New(t).Interface()
	if optVal, ok := val.(util.OptionalValue); ok {
		return b.GetTypeName(optVal.GetInnerType())
	}
	if abstr, ok := val.(openehr.UnionType); ok {
		return b.GetTypeName(abstr.GetBaseType())
	}

	if t.Kind() == reflect.Slice {
		return "[]" + b.GetTypeName(t.Elem())
	}

	if t.Kind() == reflect.Struct {
		return t.Name()
	}

	if t.Kind() == reflect.Int || t.Kind() == reflect.Int8 || t.Kind() == reflect.Int16 || t.Kind() == reflect.Int32 || t.Kind() == reflect.Int64 {
		return "int"
	}

	if t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64 {
		return "float"
	}

	if t.Kind() == reflect.String {
		return "string"
	}

	if t.Kind() == reflect.Bool {
		return "bool"
	}

	if t.Kind() == reflect.Interface {
		return "interface"
	}

	return "unknown"
}

func (b *Builder) GetFieldTypeByJSONTag(t reflect.Type, jsonTag string) (reflect.Type, bool) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() == reflect.Slice {
		return b.GetFieldTypeByJSONTag(t.Elem(), jsonTag)
	}

	if t.Kind() != reflect.Struct {
		return nil, false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// Handle comma-separated options like `json:"email_address,omitempty"`
		if tag != "" {
			tag = b.TagSplit(tag)
		}

		if tag == jsonTag {
			return field.Type, true
		}
	}

	// Check if the type is Optional[T]
	val := reflect.New(t).Interface()
	if optValue, ok := val.(util.OptionalValue); ok {
		return b.GetFieldTypeByJSONTag(optValue.GetInnerType(), jsonTag)
	}
	if abstr, ok := val.(openehr.UnionType); ok {
		return b.GetFieldTypeByJSONTag(abstr.GetBaseType(), jsonTag)
	}

	return nil, false
}

func (b *Builder) TagSplit(tag string) string {
	if idx := len(tag); idx > 0 {
		for i := 0; i < len(tag); i++ {
			if tag[i] == ',' {
				return tag[:i]
			}
		}
	}
	return tag
}
