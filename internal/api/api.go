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

// структура для передачи метрик
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// структура для передачи метрик между функциями
type MetricsMap struct {
	Metrics map[string]Metrics
}

// метод для удаления указателей в структуре Metrisc
func (metrics *Metrics) Clean() {

	if metrics.Delta != nil {
		metrics.Delta = nil
	}
	if metrics.Value != nil {
		metrics.Value = nil
	}
}
