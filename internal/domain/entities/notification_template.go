package entities

import (
	"time"
)

// NotificationTemplate plantilla de notificación
// Mapea exactamente la tabla notifications.templates
type NotificationTemplate struct {
	ID   int64  `json:"id" db:"id"`
	Code string `json:"code" db:"code"`
	Name string `json:"name" db:"name"`

	// CORREGIDO: Estos campos son JSONB en la BD, no map[string]string simple
	SubjectTranslations map[string]string `json:"subject_translations" db:"subject_translations,type:jsonb"`
	BodyTranslations    map[string]string `json:"body_translations" db:"body_translations,type:jsonb"`

	// CAMPOS FALTANTES de la tabla notifications.templates
	AvailableVariables []string `json:"available_variables,omitempty" db:"available_variables,type:text[]"`
	Channel            string   `json:"channel" db:"channel"`
	IsActive           bool     `json:"is_active" db:"is_active"`
	Priority           int      `json:"priority" db:"priority"`
	Category           string   `json:"category" db:"category"`
	Tags               []string `json:"tags,omitempty" db:"tags,type:text[]"`

	// CORREGIDO: time.Time en lugar de string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TemplateUsageStats estadísticas de uso de plantilla
type TemplateUsageStats struct {
	TotalUses     int64   `json:"total_uses"`
	SuccessUses   int64   `json:"success_uses"`
	FailureUses   int64   `json:"failure_uses"`
	AvgDeliveryMs float64 `json:"avg_delivery_ms,omitempty"`
}

// TemplateUsage uso de plantilla
type TemplateUsage struct {
	TemplateID   int64     `json:"template_id"`
	TemplateCode string    `json:"template_code"`
	UseCount     int64     `json:"use_count"`
	LastUsed     time.Time `json:"last_used,omitempty"`
}

// Métodos de utilidad para NotificationTemplate

// GetSubject obtiene el asunto en el idioma especificado
func (nt *NotificationTemplate) GetSubject(language string) string {
	if subject, ok := nt.SubjectTranslations[language]; ok && subject != "" {
		return subject
	}
	// Fallback a español
	if subject, ok := nt.SubjectTranslations["es"]; ok {
		return subject
	}
	// Fallback a cualquier idioma disponible
	for _, subject := range nt.SubjectTranslations {
		if subject != "" {
			return subject
		}
	}
	return ""
}

// GetBody obtiene el cuerpo en el idioma especificado
func (nt *NotificationTemplate) GetBody(language string) string {
	if body, ok := nt.BodyTranslations[language]; ok && body != "" {
		return body
	}
	// Fallback a español
	if body, ok := nt.BodyTranslations["es"]; ok {
		return body
	}
	// Fallback a cualquier idioma disponible
	for _, body := range nt.BodyTranslations {
		if body != "" {
			return body
		}
	}
	return ""
}

// ValidateVariables verifica que todas las variables requeridas estén presentes
func (nt *NotificationTemplate) ValidateVariables(provided map[string]interface{}) (missing []string) {
	if nt.AvailableVariables == nil {
		return nil
	}

	for _, required := range nt.AvailableVariables {
		if _, ok := provided[required]; !ok {
			missing = append(missing, required)
		}
	}
	return missing
}

// IsEmailTemplate verifica si es una plantilla de email
func (nt *NotificationTemplate) IsEmailTemplate() bool {
	return nt.Channel == "email"
}

// IsSMSTemplate verifica si es una plantilla de SMS
func (nt *NotificationTemplate) IsSMSTemplate() bool {
	return nt.Channel == "sms"
}

// IsPushTemplate verifica si es una plantilla de push notification
func (nt *NotificationTemplate) IsPushTemplate() bool {
	return nt.Channel == "push"
}

// HasTag verifica si la plantilla tiene una etiqueta específica
func (nt *NotificationTemplate) HasTag(tag string) bool {
	if nt.Tags == nil {
		return false
	}
	for _, t := range nt.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag añade una etiqueta a la plantilla
func (nt *NotificationTemplate) AddTag(tag string) {
	if nt.Tags == nil {
		nt.Tags = []string{}
	}

	// Verificar si ya existe
	for _, t := range nt.Tags {
		if t == tag {
			return
		}
	}

	nt.Tags = append(nt.Tags, tag)
}

// RemoveTag elimina una etiqueta de la plantilla
func (nt *NotificationTemplate) RemoveTag(tag string) {
	if nt.Tags == nil {
		return
	}

	newTags := []string{}
	for _, t := range nt.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	if len(newTags) == 0 {
		nt.Tags = nil
	} else {
		nt.Tags = newTags
	}
}

// GetSupportedLanguages obtiene los idiomas soportados
func (nt *NotificationTemplate) GetSupportedLanguages() []string {
	languages := make(map[string]bool)

	for lang := range nt.SubjectTranslations {
		languages[lang] = true
	}
	for lang := range nt.BodyTranslations {
		languages[lang] = true
	}

	result := make([]string, 0, len(languages))
	for lang := range languages {
		result = append(result, lang)
	}

	return result
}

// IsCompleteTranslation verifica si un idioma tiene traducción completa
func (nt *NotificationTemplate) IsCompleteTranslation(language string) bool {
	_, hasSubject := nt.SubjectTranslations[language]
	_, hasBody := nt.BodyTranslations[language]
	return hasSubject && hasBody && nt.SubjectTranslations[language] != "" && nt.BodyTranslations[language] != ""
}

// GetPriorityLevel obtiene el nivel de prioridad como string
func (nt *NotificationTemplate) GetPriorityLevel() string {
	switch {
	case nt.Priority >= 5:
		return "high"
	case nt.Priority >= 3:
		return "medium"
	default:
		return "low"
	}
}

// TemplateCategory representa las categorías predefinidas
var TemplateCategories = struct {
	General     string
	Purchase    string
	Reservation string
	Reminder    string
	Marketing   string
	Alert       string
	Security    string
}{
	General:     "general",
	Purchase:    "purchase",
	Reservation: "reservation",
	Reminder:    "reminder",
	Marketing:   "marketing",
	Alert:       "alert",
	Security:    "security",
}
