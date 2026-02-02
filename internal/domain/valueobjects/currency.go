package valueobjects

import (
	"errors"
	"strings"
)

// Currency representa una moneda ISO 4217
type Currency string

const (
	CurrencyMXN Currency = "MXN" // Peso mexicano
	CurrencyUSD Currency = "USD" // Dólar estadounidense
	CurrencyEUR Currency = "EUR" // Euro
)

// ValidCurrencies lista de monedas válidas
var ValidCurrencies = map[Currency]CurrencyInfo{
	CurrencyMXN: {Code: "MXN", Name: "Mexican Peso", Symbol: "$", DecimalPlaces: 2},
	CurrencyUSD: {Code: "USD", Name: "US Dollar", Symbol: "US$", DecimalPlaces: 2},
	CurrencyEUR: {Code: "EUR", Name: "Euro", Symbol: "€", DecimalPlaces: 2},
}

// CurrencyInfo contiene información sobre una moneda
type CurrencyInfo struct {
	Code          string
	Name          string
	Symbol        string
	DecimalPlaces int
}

// NewCurrency crea una nueva Currency validada
func NewCurrency(code string) (Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	currency := Currency(code)

	if !currency.IsValid() {
		return "", errors.New("invalid currency code")
	}

	return currency, nil
}

// IsValid verifica si la moneda es válida
func (c Currency) IsValid() bool {
	_, exists := ValidCurrencies[c]
	return exists
}

// Code devuelve el código ISO
func (c Currency) Code() string {
	return string(c)
}

// Name devuelve el nombre de la moneda
func (c Currency) Name() string {
	if info, exists := ValidCurrencies[c]; exists {
		return info.Name
	}
	return string(c)
}

// Symbol devuelve el símbolo de la moneda
func (c Currency) Symbol() string {
	if info, exists := ValidCurrencies[c]; exists {
		return info.Symbol
	}
	return string(c)
}

// DecimalPlaces devuelve el número de decimales
func (c Currency) DecimalPlaces() int {
	if info, exists := ValidCurrencies[c]; exists {
		return info.DecimalPlaces
	}
	return 2 // Por defecto
}

// Equals compara dos monedas
func (c Currency) Equals(other Currency) bool {
	return c == other
}

// String devuelve el código como string
func (c Currency) String() string {
	return string(c)
}

// GetDefaultCurrency devuelve la moneda por defecto
func GetDefaultCurrency() Currency {
	return CurrencyMXN
}

// GetAllCurrencies devuelve todas las monedas válidas
func GetAllCurrencies() []Currency {
	currencies := make([]Currency, 0, len(ValidCurrencies))
	for currency := range ValidCurrencies {
		currencies = append(currencies, currency)
	}
	return currencies
}

// GetCurrencyInfo devuelve la información de una moneda
func GetCurrencyInfo(code string) (CurrencyInfo, error) {
	currency, err := NewCurrency(code)
	if err != nil {
		return CurrencyInfo{}, err
	}

	info, exists := ValidCurrencies[currency]
	if !exists {
		return CurrencyInfo{}, errors.New("currency not found")
	}

	return info, nil
}

// SupportsCurrency verifica si una moneda está soportada
func SupportsCurrency(code string) bool {
	currency, err := NewCurrency(code)
	if err != nil {
		return false
	}
	return currency.IsValid()
}

// ParseCurrency intenta parsear un string a Currency
func ParseCurrency(s string) (Currency, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return GetDefaultCurrency(), nil
	}

	// Intentar por código
	if currency, err := NewCurrency(s); err == nil {
		return currency, nil
	}

	// Intentar por símbolo
	for currency, info := range ValidCurrencies {
		if info.Symbol == s {
			return currency, nil
		}
	}

	// Intentar por nombre (case insensitive)
	sLower := strings.ToLower(s)
	for currency, info := range ValidCurrencies {
		if strings.ToLower(info.Name) == sLower {
			return currency, nil
		}
	}

	return "", errors.New("unable to parse currency")
}
