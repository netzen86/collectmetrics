package api

const (
	Th           string = "text"
	HTML         string = "text/html"
	Js           string = "application/json"
	Gz           string = "gzip"
	TemplatePath string = "/web/template/metrics.html"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
