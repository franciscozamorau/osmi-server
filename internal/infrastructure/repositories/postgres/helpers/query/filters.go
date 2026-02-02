package query

import (
	"fmt"
	"strings"
	"time"
)

// FilterOperator representa un operador de filtro
type FilterOperator string

const (
	OpEquals         FilterOperator = "="
	OpNotEquals      FilterOperator = "!="
	OpGreaterThan    FilterOperator = ">"
	OpGreaterOrEqual FilterOperator = ">="
	OpLessThan       FilterOperator = "<"
	OpLessOrEqual    FilterOperator = "<="
	OpLike           FilterOperator = "LIKE"
	OpILike          FilterOperator = "ILIKE"
	OpIn             FilterOperator = "IN"
	OpNotIn          FilterOperator = "NOT IN"
	OpIsNull         FilterOperator = "IS NULL"
	OpIsNotNull      FilterOperator = "IS NOT NULL"
	OpBetween        FilterOperator = "BETWEEN"
)

// Filter representa un filtro individual
type Filter struct {
	Field    string
	Operator FilterOperator
	Value    interface{}
	Values   []interface{} // Para operadores IN, NOT IN
	Start    interface{}   // Para BETWEEN
	End      interface{}   // Para BETWEEN
}

// NewFilter crea un nuevo filtro
func NewFilter(field string, operator FilterOperator, value interface{}) *Filter {
	return &Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// NewInFilter crea un filtro IN
func NewInFilter(field string, values []interface{}) *Filter {
	return &Filter{
		Field:    field,
		Operator: OpIn,
		Values:   values,
	}
}

// NewBetweenFilter crea un filtro BETWEEN
func NewBetweenFilter(field string, start, end interface{}) *Filter {
	return &Filter{
		Field:    field,
		Operator: OpBetween,
		Start:    start,
		End:      end,
	}
}

// FilterBuilder construye condiciones de filtro
type FilterBuilder struct {
	filters []*Filter
	args    []interface{}
	counter int
}

// NewFilterBuilder crea un nuevo FilterBuilder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		filters: make([]*Filter, 0),
		args:    make([]interface{}, 0),
		counter: 1,
	}
}

// AddFilter añade un filtro
func (fb *FilterBuilder) AddFilter(filter *Filter) *FilterBuilder {
	fb.filters = append(fb.filters, filter)
	return fb
}

// AddFilters añade múltiples filtros
func (fb *FilterBuilder) AddFilters(filters ...*Filter) *FilterBuilder {
	fb.filters = append(fb.filters, filters...)
	return fb
}

// Build construye las condiciones WHERE
func (fb *FilterBuilder) Build() (string, []interface{}) {
	if len(fb.filters) == 0 {
		return "", nil
	}

	var conditions []string
	fb.counter = 1
	fb.args = make([]interface{}, 0)

	for _, filter := range fb.filters {
		condition, args := fb.buildFilter(filter)
		if condition != "" {
			conditions = append(conditions, condition)
			fb.args = append(fb.args, args...)
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return strings.Join(conditions, " AND "), fb.args
}

// buildFilter construye una condición individual
func (fb *FilterBuilder) buildFilter(filter *Filter) (string, []interface{}) {
	switch filter.Operator {
	case OpIsNull:
		return filter.Field + " IS NULL", nil
	case OpIsNotNull:
		return filter.Field + " IS NOT NULL", nil
	case OpIn:
		if len(filter.Values) == 0 {
			return "1 = 0", nil // Nunca coincidirá
		}
		placeholders := make([]string, len(filter.Values))
		args := make([]interface{}, len(filter.Values))
		for i, value := range filter.Values {
			placeholders[i] = fmt.Sprintf("$%d", fb.counter)
			args[i] = value
			fb.counter++
		}
		return fmt.Sprintf("%s IN (%s)", filter.Field, strings.Join(placeholders, ", ")), args
	case OpNotIn:
		if len(filter.Values) == 0 {
			return "1 = 1", nil // Siempre coincidirá
		}
		placeholders := make([]string, len(filter.Values))
		args := make([]interface{}, len(filter.Values))
		for i, value := range filter.Values {
			placeholders[i] = fmt.Sprintf("$%d", fb.counter)
			args[i] = value
			fb.counter++
		}
		return fmt.Sprintf("%s NOT IN (%s)", filter.Field, strings.Join(placeholders, ", ")), args
	case OpBetween:
		condition := fmt.Sprintf("%s BETWEEN $%d AND $%d", filter.Field, fb.counter, fb.counter+1)
		args := []interface{}{filter.Start, filter.End}
		fb.counter += 2
		return condition, args
	default:
		condition := fmt.Sprintf("%s %s $%d", filter.Field, filter.Operator, fb.counter)
		args := []interface{}{filter.Value}
		fb.counter++
		return condition, args
	}
}

// TextFilter crea un filtro de texto
func TextFilter(field, value string, exact bool) *Filter {
	if exact {
		return NewFilter(field, OpEquals, value)
	}
	return NewFilter(field, OpILike, "%"+value+"%")
}

// DateFilter crea un filtro de fecha
func DateFilter(field string, date time.Time) *Filter {
	return NewFilter(field, OpEquals, date)
}

// DateRangeFilter crea un filtro de rango de fechas
func DateRangeFilter(field string, start, end time.Time) *Filter {
	return NewBetweenFilter(field, start, end)
}

// NumberFilter crea un filtro numérico
func NumberFilter(field string, value interface{}, operator FilterOperator) *Filter {
	return NewFilter(field, operator, value)
}

// BooleanFilter crea un filtro booleano
func BooleanFilter(field string, value bool) *Filter {
	return NewFilter(field, OpEquals, value)
}

// NullFilter crea un filtro IS NULL
func NullFilter(field string) *Filter {
	return &Filter{
		Field:    field,
		Operator: OpIsNull,
	}
}

// NotNullFilter crea un filtro IS NOT NULL
func NotNullFilter(field string) *Filter {
	return &Filter{
		Field:    field,
		Operator: OpIsNotNull,
	}
}
