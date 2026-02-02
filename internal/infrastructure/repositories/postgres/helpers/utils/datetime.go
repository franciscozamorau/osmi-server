package utils

import (
	"fmt"
	"time"
)

// FormatDateForDB formatea fecha para consultas SQL
func FormatDateForDB(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDateISO formatea fecha en formato ISO 8601
func FormatDateISO(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatDatePretty formatea fecha de forma legible
func FormatDatePretty(t time.Time) string {
	return t.Format("02 Jan 2006 15:04")
}

// FormatDateShort formatea fecha corta
func FormatDateShort(t time.Time) string {
	return t.Format("02/01/2006")
}

// FormatTimeOnly formatea solo hora
func FormatTimeOnly(t time.Time) string {
	return t.Format("15:04")
}

// ParseDateFromString parsea fecha de string con múltiples formatos
func ParseDateFromString(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		"02-01-2006",
		"January 2, 2006",
		"2006-01-02T15:04:05",
		"20060102",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// BeginningOfDay obtiene el inicio del día
func BeginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay obtiene el fin del día
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// BeginningOfWeek obtiene inicio de semana (lunes)
func BeginningOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return t.AddDate(0, 0, -int(weekday)+1)
}

// EndOfWeek obtiene fin de semana (domingo)
func EndOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		return EndOfDay(t)
	}
	return EndOfDay(t.AddDate(0, 0, 7-int(weekday)))
}

// BeginningOfMonth obtiene inicio de mes
func BeginningOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth obtiene fin de mes
func EndOfMonth(t time.Time) time.Time {
	return BeginningOfMonth(t).AddDate(0, 1, -1)
}

// BeginningOfYear obtiene inicio de año
func BeginningOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear obtiene fin de año
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// IsToday verifica si una fecha es hoy
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsYesterday verifica si una fecha es ayer
func IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}

// IsTomorrow verifica si una fecha es mañana
func IsTomorrow(t time.Time) bool {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return t.Year() == tomorrow.Year() && t.Month() == tomorrow.Month() && t.Day() == tomorrow.Day()
}

// IsThisWeek verifica si una fecha es esta semana
func IsThisWeek(t time.Time) bool {
	now := time.Now()
	year, week := now.ISOWeek()
	tYear, tWeek := t.ISOWeek()
	return year == tYear && week == tWeek
}

// IsThisMonth verifica si una fecha es este mes
func IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}

// IsThisYear verifica si una fecha es este año
func IsThisYear(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year()
}

// DaysBetween calcula días entre dos fechas
func DaysBetween(start, end time.Time) int {
	start = BeginningOfDay(start)
	end = BeginningOfDay(end)
	hours := end.Sub(start).Hours()
	return int(hours / 24)
}

// HoursBetween calcula horas entre dos fechas
func HoursBetween(start, end time.Time) int {
	hours := end.Sub(start).Hours()
	return int(hours)
}

// MinutesBetween calcula minutos entre dos fechas
func MinutesBetween(start, end time.Time) int {
	minutes := end.Sub(start).Minutes()
	return int(minutes)
}

// WeeksBetween calcula semanas entre dos fechas
func WeeksBetween(start, end time.Time) int {
	days := DaysBetween(start, end)
	return days / 7
}

// MonthsBetween calcula meses entre dos fechas
func MonthsBetween(start, end time.Time) int {
	years := end.Year() - start.Year()
	months := int(end.Month()) - int(start.Month())

	if end.Day() < start.Day() {
		months--
	}

	return years*12 + months
}

// YearsBetween calcula años entre dos fechas
func YearsBetween(start, end time.Time) int {
	years := end.Year() - start.Year()

	// Ajustar si el cumpleaños aún no ha pasado este año
	if end.Month() < start.Month() || (end.Month() == start.Month() && end.Day() < start.Day()) {
		years--
	}

	return years
}

// AddBusinessDays añade días hábiles (excluye fines de semana)
func AddBusinessDays(t time.Time, days int) time.Time {
	result := t
	added := 0

	for added < days {
		result = result.AddDate(0, 0, 1)
		if result.Weekday() != time.Saturday && result.Weekday() != time.Sunday {
			added++
		}
	}

	return result
}

// IsBusinessDay verifica si es día hábil
func IsBusinessDay(t time.Time) bool {
	return t.Weekday() != time.Saturday && t.Weekday() != time.Sunday
}

// IsWeekend verifica si es fin de semana
func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

// IsLeapYear verifica si es año bisiesto
func IsLeapYear(year int) bool {
	return year%400 == 0 || (year%4 == 0 && year%100 != 0)
}

// Age calcula edad a partir de fecha de nacimiento
func Age(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// Ajustar si el cumpleaños aún no ha pasado este año
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	return age
}

// IsValidDateRange verifica rango de fechas válido
func IsValidDateRange(start, end time.Time) bool {
	return !start.IsZero() && !end.IsZero() && end.After(start)
}

// MinDate devuelve la fecha más temprana
func MinDate(dates ...time.Time) time.Time {
	if len(dates) == 0 {
		return time.Time{}
	}

	min := dates[0]
	for _, date := range dates[1:] {
		if date.Before(min) {
			min = date
		}
	}
	return min
}

// MaxDate devuelve la fecha más tardía
func MaxDate(dates ...time.Time) time.Time {
	if len(dates) == 0 {
		return time.Time{}
	}

	max := dates[0]
	for _, date := range dates[1:] {
		if date.After(max) {
			max = date
		}
	}
	return max
}

// ParseDuration parsea duración de string
func ParseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

// FormatDuration formatea duración de forma legible
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%d hours %d minutes", hours, minutes)
}

// TimeInTimezone convierte tiempo a zona horaria específica
func TimeInTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

// NowInTimezone obtiene hora actual en zona horaria específica
func NowInTimezone(timezone string) (time.Time, error) {
	return TimeInTimezone(time.Now(), timezone)
}

// IsDateInFuture verifica si fecha está en futuro
func IsDateInFuture(t time.Time) bool {
	return t.After(time.Now())
}

// IsDateInPast verifica si fecha está en pasado
func IsDateInPast(t time.Time) bool {
	return t.Before(time.Now())
}

// RoundToNearestMinute redondea al minuto más cercano
func RoundToNearestMinute(t time.Time) time.Time {
	seconds := t.Second()
	nanos := t.Nanosecond()

	if seconds >= 30 || (seconds == 29 && nanos > 500000000) {
		return t.Add(time.Minute).Truncate(time.Minute)
	}

	return t.Truncate(time.Minute)
}

// RoundToNearestHour redondea a la hora más cercana
func RoundToNearestHour(t time.Time) time.Time {
	minutes := t.Minute()
	seconds := t.Second()
	nanos := t.Nanosecond()

	if minutes >= 30 || (minutes == 29 && (seconds > 0 || nanos > 0)) {
		return t.Add(time.Hour).Truncate(time.Hour)
	}

	return t.Truncate(time.Hour)
}

// GenerateDateRange genera rango de fechas
func GenerateDateRange(start, end time.Time, interval time.Duration) []time.Time {
	var dates []time.Time

	for current := start; !current.After(end); current = current.Add(interval) {
		dates = append(dates, current)
	}

	return dates
}

// IsDateBetween verifica si fecha está entre otras dos
func IsDateBetween(date, start, end time.Time) bool {
	return (date.Equal(start) || date.After(start)) && (date.Equal(end) || date.Before(end))
}
