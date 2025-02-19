package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Definição de métricas do Prometheus

var (
	// paymentsTotal é um contador que rastreia o número total de pagamentos processados,
	// categorizados pelo status (ex: "success", "failed").
	paymentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ecommerce_payments_total", // Nome da métrica no Prometheus
			Help: "Total number of payments processed",
		},
		[]string{"status"}, // Rótulo para diferenciar pagamentos bem-sucedidos ou falhos
	)

	// httpDuration é um histograma que mede a duração das requisições HTTP,
	// categorizadas pelo caminho do handler.
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ecommerce_http_duration",
			Help:    "HTTP duration in seconds",
			Buckets: prometheus.DefBuckets, // Usa os buckets padrão do Prometheus
		},
		[]string{"handler"}, // Rótulo para identificar qual rota está sendo monitorada
	)
)

// Middleware para medir o tempo de execução das requisições HTTP
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inicia o cronômetro para medir a duração da requisição
		duration := prometheus.NewTimer(httpDuration.WithLabelValues(r.URL.Path))
		next.ServeHTTP(w, r)
		// Registra a duração no histograma
		duration.ObserveDuration()
	})
}

// init é chamado antes de main para registrar as métricas no Prometheus
func init() {
	prometheus.MustRegister(paymentsTotal, httpDuration)
}

func main() {
	// Criando um roteador HTTP
	mux := http.NewServeMux()

	// Rota para processar pagamentos
	mux.Handle("POST /payments", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := "success"

		// Verifica se o corpo da requisição pode ser processado
		if err := r.ParseForm(); err != nil {
			http.Error(w, "unprocessable entity", http.StatusUnprocessableEntity)
			return
		}

		// Se o status enviado não for "success", considera como "failed"
		if r.PostForm.Get("status") != "success" {
			status = "failed"
		}

		// Incrementa o contador de pagamentos com base no status
		paymentsTotal.WithLabelValues(status).Inc()
		w.Write([]byte("payments requested"))
	})))

	// Rota para expor as métricas do Prometheus
	mux.Handle("GET /metrics", promhttp.Handler())

	// Inicia o servidor na porta 8080
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
