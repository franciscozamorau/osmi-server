package valueobjects

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Money representa una cantidad monetaria con moneda
type Money struct {
	amount   int64    // Cantidad en la unidad más pequeña (centavos)
	currency Currency // Moneda
}

// NewMoney crea un nuevo Money
func NewMoney(amount float64, currency Currency) (Money, error) {
	if !currency.IsValid() {
		return Money{}, errors.New("invalid currency")
	}

	// Convertir a la unidad más pequeña
	minorUnits := currency.DecimalPlaces()
	amountInMinor := int64(math.Round(amount * math.Pow10(minorUnits)))

	return Money{
		amount:   amountInMinor,
		currency: currency,
	}, nil
}

// NewMoneyFromMinor crea Money desde la unidad más pequeña
func NewMoneyFromMinor(amount int64, currency Currency) (Money, error) {
	if !currency.IsValid() {
		return Money{}, errors.New("invalid currency")
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

// Amount devuelve la cantidad como float64
func (m Money) Amount() float64 {
	minorUnits := m.currency.DecimalPlaces()
	return float64(m.amount) / math.Pow10(minorUnits)
}

// AmountInMinor devuelve la cantidad en la unidad más pequeña
func (m Money) AmountInMinor() int64 {
	return m.amount
}

// Currency devuelve la moneda
func (m Money) Currency() Currency {
	return m.currency
}

// String devuelve la representación string formateada
func (m Money) String() string {
	amount := m.Amount()
	symbol := m.currency.Symbol()

	// Formatear según las reglas de la moneda
	switch m.currency {
	case CurrencyUSD, CurrencyEUR:
		return fmt.Sprintf("%s%.2f", symbol, amount)
	case CurrencyMXN:
		return fmt.Sprintf("$%.2f", amount) // Peso mexicano
	default:
		return fmt.Sprintf("%s %.2f", m.currency.Code(), amount)
	}
}

// Add suma dos cantidades de la misma moneda
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot add money with different currencies")
	}

	return Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

// Subtract resta dos cantidades de la misma moneda
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot subtract money with different currencies")
	}

	if m.amount < other.amount {
		return Money{}, errors.New("insufficient funds")
	}

	return Money{
		amount:   m.amount - other.amount,
		currency: m.currency,
	}, nil
}

// Multiply multiplica por un escalar
func (m Money) Multiply(scalar float64) Money {
	result := float64(m.amount) * scalar
	return Money{
		amount:   int64(math.Round(result)),
		currency: m.currency,
	}
}

// Divide divide por un escalar
func (m Money) Divide(scalar float64) (Money, error) {
	if scalar == 0 {
		return Money{}, errors.New("division by zero")
	}

	result := float64(m.amount) / scalar
	return Money{
		amount:   int64(math.Round(result)),
		currency: m.currency,
	}, nil
}

// Percentage calcula un porcentaje
func (m Money) Percentage(percent float64) Money {
	if percent < 0 || percent > 100 {
		percent = math.Max(0, math.Min(100, percent))
	}

	amount := float64(m.amount) * (percent / 100)
	return Money{
		amount:   int64(math.Round(amount)),
		currency: m.currency,
	}
}

// IsZero verifica si la cantidad es cero
func (m Money) IsZero() bool {
	return m.amount == 0
}

// IsPositive verifica si la cantidad es positiva
func (m Money) IsPositive() bool {
	return m.amount > 0
}

// IsNegative verifica si la cantidad es negativa
func (m Money) IsNegative() bool {
	return m.amount < 0
}

// Equals compara dos cantidades de dinero
func (m Money) Equals(other Money) bool {
	return m.currency == other.currency && m.amount == other.amount
}

// GreaterThan verifica si es mayor que otra cantidad
func (m Money) GreaterThan(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, errors.New("cannot compare money with different currencies")
	}
	return m.amount > other.amount, nil
}

// LessThan verifica si es menor que otra cantidad
func (m Money) LessThan(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, errors.New("cannot compare money with different currencies")
	}
	return m.amount < other.amount, nil
}

// ParseMoney parsea un string a Money
func ParseMoney(s string, defaultCurrency Currency) (Money, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return NewMoney(0, defaultCurrency)
	}

	// Buscar símbolo de moneda o código
	var currency Currency = defaultCurrency

	// Intentar detectar la moneda
	for _, c := range []Currency{CurrencyUSD, CurrencyEUR, CurrencyMXN} {
		symbol := c.Symbol()
		code := c.Code()

		if strings.HasPrefix(s, symbol) {
			currency = c
			s = strings.TrimPrefix(s, symbol)
			break
		} else if strings.HasPrefix(strings.ToUpper(s), code+" ") {
			currency = c
			s = strings.TrimPrefix(s, code+" ")
			break
		}
	}

	// Limpiar separadores de miles
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")

	// Parsear el número
	amount, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Money{}, fmt.Errorf("invalid amount: %v", err)
	}

	return NewMoney(amount, currency)
}

// FormatWithCurrency formatea con el código de moneda
func (m Money) FormatWithCurrency() string {
	return fmt.Sprintf("%.2f %s", m.Amount(), m.currency.Code())
}

// FormatLocalized formatea según las reglas locales
func (m Money) FormatLocalized(locale string) string {
	// Por ahora, usamos el formato básico
	// En una implementación real, usaríamos un paquete de internacionalización
	return m.String()
}

// Split divide en partes iguales
func (m Money) Split(parts int) ([]Money, error) {
	if parts <= 0 {
		return nil, errors.New("parts must be positive")
	}

	amountPerPart := m.amount / int64(parts)
	remainder := m.amount % int64(parts)

	result := make([]Money, parts)
	for i := 0; i < parts; i++ {
		amount := amountPerPart
		if int64(i) < remainder {
			amount++
		}
		result[i] = Money{
			amount:   amount,
			currency: m.currency,
		}
	}

	return result, nil
}

// Allocate distribuye según porcentajes
func (m Money) Allocate(percentages []float64) ([]Money, error) {
	var totalPercent float64
	for _, p := range percentages {
		if p < 0 {
			return nil, errors.New("percentages cannot be negative")
		}
		totalPercent += p
	}

	if math.Abs(totalPercent-100) > 0.01 {
		return nil, errors.New("percentages must sum to 100")
	}

	result := make([]Money, len(percentages))
	allocated := int64(0)

	for i, p := range percentages {
		amount := int64(math.Round(float64(m.amount) * p / 100))
		result[i] = Money{
			amount:   amount,
			currency: m.currency,
		}
		allocated += amount
	}

	// Ajustar por errores de redondeo
	adjustment := m.amount - allocated
	if adjustment != 0 {
		// Añadir el ajuste al último elemento
		lastIdx := len(result) - 1
		result[lastIdx].amount += adjustment
	}

	return result, nil
}
