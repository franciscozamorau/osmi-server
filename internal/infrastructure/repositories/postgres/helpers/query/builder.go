package query

import (
	"fmt"
	"strings"
	"time"
)

// QueryBuilder construye queries SQL de forma segura
type QueryBuilder struct {
	query      strings.Builder
	args       []interface{}
	argCounter int
	conditions []string
	joins      []string
	orderBy    []string
	groupBy    []string
	having     []string
	distinct   bool
	limit      int
	offset     int
}

// NewQueryBuilder crea un nuevo QueryBuilder
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	qb := &QueryBuilder{
		args:       make([]interface{}, 0),
		argCounter: 1,
		limit:      -1,
		offset:     -1,
	}
	qb.query.WriteString(baseQuery)
	return qb
}

// Select inicia una query SELECT
func Select(columns ...string) *QueryBuilder {
	cols := "*"
	if len(columns) > 0 {
		cols = strings.Join(columns, ", ")
	}
	return NewQueryBuilder("SELECT " + cols)
}

// From especifica la tabla
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.query.WriteString(" FROM " + table)
	return qb
}

// Distinct aplica DISTINCT
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.distinct = true
	queryStr := qb.query.String()
	qb.query.Reset()
	qb.query.WriteString(strings.Replace(queryStr, "SELECT", "SELECT DISTINCT", 1))
	return qb
}

// Where añade una condición WHERE
func (qb *QueryBuilder) Where(condition string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, value)
	qb.argCounter++
	return qb
}

// WhereIn añade condición WHERE IN
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb.Where("1 = 0") // Nunca coincidirá
	}

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = fmt.Sprintf("$%d", qb.argCounter)
		qb.args = append(qb.args, values[i])
		qb.argCounter++
	}

	condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereNotIn añade condición WHERE NOT IN
func (qb *QueryBuilder) WhereNotIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb.Where("1 = 1") // Siempre coincidirá
	}

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = fmt.Sprintf("$%d", qb.argCounter)
		qb.args = append(qb.args, values[i])
		qb.argCounter++
	}

	condition := fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", "))
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereLike añade condición WHERE LIKE
func (qb *QueryBuilder) WhereLike(field, value string, caseSensitive bool) *QueryBuilder {
	operator := "LIKE"
	if !caseSensitive {
		operator = "ILIKE"
	}
	condition := fmt.Sprintf("%s %s $%d", field, operator, qb.argCounter)
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, "%"+value+"%")
	qb.argCounter++
	return qb
}

// WhereBetween añade condición WHERE BETWEEN
func (qb *QueryBuilder) WhereBetween(field string, start, end interface{}) *QueryBuilder {
	condition := fmt.Sprintf("%s BETWEEN $%d AND $%d", field, qb.argCounter, qb.argCounter+1)
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, start, end)
	qb.argCounter += 2
	return qb
}

// WhereDateBetween añade condición WHERE para rango de fechas
func (qb *QueryBuilder) WhereDateBetween(field string, start, end time.Time) *QueryBuilder {
	return qb.WhereBetween(field, start, end)
}

// WhereIsNull añade condición WHERE IS NULL
func (qb *QueryBuilder) WhereIsNull(field string) *QueryBuilder {
	qb.conditions = append(qb.conditions, field+" IS NULL")
	return qb
}

// WhereIsNotNull añade condición WHERE IS NOT NULL
func (qb *QueryBuilder) WhereIsNotNull(field string) *QueryBuilder {
	qb.conditions = append(qb.conditions, field+" IS NOT NULL")
	return qb
}

// Join añade un JOIN
func (qb *QueryBuilder) Join(table, on string) *QueryBuilder {
	qb.joins = append(qb.joins, "JOIN "+table+" ON "+on)
	return qb
}

// LeftJoin añade un LEFT JOIN
func (qb *QueryBuilder) LeftJoin(table, on string) *QueryBuilder {
	qb.joins = append(qb.joins, "LEFT JOIN "+table+" ON "+on)
	return qb
}

// RightJoin añade un RIGHT JOIN
func (qb *QueryBuilder) RightJoin(table, on string) *QueryBuilder {
	qb.joins = append(qb.joins, "RIGHT JOIN "+table+" ON "+on)
	return qb
}

// InnerJoin añade un INNER JOIN
func (qb *QueryBuilder) InnerJoin(table, on string) *QueryBuilder {
	qb.joins = append(qb.joins, "INNER JOIN "+table+" ON "+on)
	return qb
}

// OrderBy añade ORDER BY
func (qb *QueryBuilder) OrderBy(field string, descending bool) *QueryBuilder {
	order := "ASC"
	if descending {
		order = "DESC"
	}
	qb.orderBy = append(qb.orderBy, field+" "+order)
	return qb
}

// OrderByRaw añade ORDER BY con expresión cruda
func (qb *QueryBuilder) OrderByRaw(expression string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, expression)
	return qb
}

// GroupBy añade GROUP BY
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, fields...)
	return qb
}

// Having añade condición HAVING
func (qb *QueryBuilder) Having(condition string, value interface{}) *QueryBuilder {
	qb.having = append(qb.having, condition)
	qb.args = append(qb.args, value)
	qb.argCounter++
	return qb
}

// Limit añade LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset añade OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Build construye la query completa
func (qb *QueryBuilder) Build() (string, []interface{}) {
	// Añadir JOINs
	for _, join := range qb.joins {
		qb.query.WriteString(" " + join)
	}

	// Añadir WHERE
	if len(qb.conditions) > 0 {
		qb.query.WriteString(" WHERE " + strings.Join(qb.conditions, " AND "))
	}

	// Añadir GROUP BY
	if len(qb.groupBy) > 0 {
		qb.query.WriteString(" GROUP BY " + strings.Join(qb.groupBy, ", "))
	}

	// Añadir HAVING
	if len(qb.having) > 0 {
		qb.query.WriteString(" HAVING " + strings.Join(qb.having, " AND "))
	}

	// Añadir ORDER BY
	if len(qb.orderBy) > 0 {
		qb.query.WriteString(" ORDER BY " + strings.Join(qb.orderBy, ", "))
	}

	// Añadir LIMIT
	if qb.limit >= 0 {
		qb.query.WriteString(fmt.Sprintf(" LIMIT $%d", qb.argCounter))
		qb.args = append(qb.args, qb.limit)
		qb.argCounter++
	}

	// Añadir OFFSET
	if qb.offset >= 0 {
		qb.query.WriteString(fmt.Sprintf(" OFFSET $%d", qb.argCounter))
		qb.args = append(qb.args, qb.offset)
		qb.argCounter++
	}

	return qb.query.String(), qb.args
}

// BuildCount construye query COUNT
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	// Extraer la parte FROM de la query original
	queryStr := qb.query.String()

	// Encontrar la posición de FROM
	fromIndex := strings.Index(strings.ToUpper(queryStr), " FROM ")
	if fromIndex == -1 {
		return "SELECT COUNT(*) FROM (" + queryStr + ") AS count_query", qb.args
	}

	// Construir query COUNT
	countQuery := "SELECT COUNT(*) " + queryStr[fromIndex:]

	// Quitar ORDER BY, LIMIT, OFFSET para COUNT
	countQuery = removeClause(countQuery, "ORDER BY")
	countQuery = removeClause(countQuery, "LIMIT")
	countQuery = removeClause(countQuery, "OFFSET")

	return countQuery, qb.args
}

// removeClause remueve una cláusula de la query
func removeClause(query, clause string) string {
	upperQuery := strings.ToUpper(query)
	upperClause := strings.ToUpper(clause)

	if idx := strings.Index(upperQuery, upperClause); idx != -1 {
		return query[:idx]
	}
	return query
}
