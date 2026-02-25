// internal/infrastructure/repositories/postgres/helpers/query/builder.go
package query

import (
	"fmt"
	"strings"
)

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

// Where - VERSIÓN CORREGIDA con soporte para múltiples placeholders
func (qb *QueryBuilder) Where(condition string, values ...interface{}) *QueryBuilder {
	// Contar placeholders en la condición ($1, $2, etc.)
	placeholderCount := strings.Count(condition, "?")

	if placeholderCount == 0 {
		// Sin placeholders, usar como raw
		qb.conditions = append(qb.conditions, condition)
		return qb
	}

	// Validar que tenemos suficientes valores
	if len(values) != placeholderCount {
		// Si no hay suficientes valores, asumir que condition ya tiene placeholders con $
		qb.conditions = append(qb.conditions, condition)
		for _, value := range values {
			qb.args = append(qb.args, value)
		}
		qb.argCounter += len(values)
		return qb
	}

	// Reemplazar ? con $n
	processedCondition := condition
	for i := 0; i < placeholderCount; i++ {
		processedCondition = strings.Replace(processedCondition, "?", fmt.Sprintf("$%d", qb.argCounter+i), 1)
	}

	qb.conditions = append(qb.conditions, processedCondition)
	qb.args = append(qb.args, values...)
	qb.argCounter += placeholderCount

	return qb
}

// WhereRaw añade condición WHERE cruda
func (qb *QueryBuilder) WhereRaw(condition string) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereIn añade condición WHERE IN
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb.WhereRaw("1 = 0")
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

// Join añade JOIN
func (qb *QueryBuilder) Join(join string) *QueryBuilder {
	qb.joins = append(qb.joins, join)
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

// Having añade HAVING
func (qb *QueryBuilder) Having(condition string, values ...interface{}) *QueryBuilder {
	placeholderCount := strings.Count(condition, "?")

	if placeholderCount == 0 {
		qb.having = append(qb.having, condition)
		return qb
	}

	processedCondition := condition
	for i := 0; i < placeholderCount; i++ {
		processedCondition = strings.Replace(processedCondition, "?", fmt.Sprintf("$%d", qb.argCounter+i), 1)
	}

	qb.having = append(qb.having, processedCondition)
	qb.args = append(qb.args, values...)
	qb.argCounter += placeholderCount

	return qb
}

// Distinct añade DISTINCT
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.distinct = true
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
	var query strings.Builder
	query.WriteString(qb.query.String())

	// Añadir JOINs
	for _, join := range qb.joins {
		query.WriteString(" " + join)
	}

	// Añadir WHERE
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE " + strings.Join(qb.conditions, " AND "))
	}

	// Añadir GROUP BY
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY " + strings.Join(qb.groupBy, ", "))
	}

	// Añadir HAVING
	if len(qb.having) > 0 {
		query.WriteString(" HAVING " + strings.Join(qb.having, " AND "))
	}

	// Añadir ORDER BY
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY " + strings.Join(qb.orderBy, ", "))
	}

	// Añadir LIMIT
	if qb.limit >= 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// Añadir OFFSET
	if qb.offset >= 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), qb.args
}

// BuildCount construye query COUNT
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
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

	if idx := strings.Index(upperQuery, " "+upperClause); idx != -1 {
		// Encontrar el final de la cláusula
		endIdx := strings.Index(upperQuery[idx+1:], " ")
		if endIdx == -1 {
			return query[:idx]
		}
		return query[:idx] + query[idx+endIdx+2:]
	}
	return query
}

// Reset resetea el builder
func (qb *QueryBuilder) Reset() {
	qb.query.Reset()
	qb.args = make([]interface{}, 0)
	qb.argCounter = 1
	qb.conditions = make([]string, 0)
	qb.joins = make([]string, 0)
	qb.orderBy = make([]string, 0)
	qb.groupBy = make([]string, 0)
	qb.having = make([]string, 0)
	qb.distinct = false
	qb.limit = -1
	qb.offset = -1
}
