// Package api - пакет содержит общие константы и структуры общие для сервера и агента
package api

// константы с типом контернта, и типом метрик
const (
	Th           string = "text"
	HTML         string = "text/html"
	Js           string = "application/json"
	Gz           string = "gzip"
	TemplatePath string = "/web/template/metrics.html"
	Gauge        string = "gauge"
	Counter      string = "counter"
)

// Metrics структура для передачи метрик
type Metrics struct {
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

// MetricsMap структура для передачи метрик между функциями
type MetricsMap struct {
	Metrics map[string]Metrics
}

// Clean метод для удаления указателей в структуре Metrisc
func (metrics *Metrics) Clean() {

	if metrics.Delta != nil {
		metrics.Delta = nil
	}
	if metrics.Value != nil {
		metrics.Value = nil
	}
}
