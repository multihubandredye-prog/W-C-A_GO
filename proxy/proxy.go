package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	WEBHOOK_PATH = "/Tasker"
	HOSTNAME     = "127.0.0.1"
	DELAY_MS     = 1000
)

type DadosProxy struct {
	RecebidoEm  time.Time       `json:"recebido_em"`
	Origem      string          `json:"origem"`
	Metodo      string          `json:"metodo"`
	URLOriginal string          `json:"url_original"`
	Headers     http.Header     `json:"headers"`
	Body        json.RawMessage `json:"body"`
}

func processadorFila(fila chan DadosProxy, portaDestino string) {
	for dados := range fila {
		time.Sleep(time.Duration(DELAY_MS) * time.Millisecond)
		payloadJSON, _ := json.Marshal(dados)
		destinoURL := fmt.Sprintf("http://%s:%s%s", HOSTNAME, portaDestino, WEBHOOK_PATH)
		proxyReq, _ := http.NewRequest("POST", destinoURL, bytes.NewBuffer(payloadJSON))
		proxyReq.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 5 * time.Second}
		client.Do(proxyReq)
	}
}

func novoHandler(fila chan DadosProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != WEBHOOK_PATH {
			http.NotFound(w, r)
			return
		}
		bodyBytes, _ := io.ReadAll(r.Body)
		var rawBody json.RawMessage
		json.Unmarshal(bodyBytes, &rawBody)
		dados := DadosProxy{
			RecebidoEm:  time.Now(),
			Origem:      "tasker",
			Metodo:      r.Method,
			URLOriginal: r.URL.Path,
			Headers:     r.Header,
			Body:        rawBody,
		}
		fila <- dados
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
	}
}

func iniciarServidor(portaOrigem string, fila chan DadosProxy) {
	mux := http.NewServeMux()
	mux.HandleFunc(WEBHOOK_PATH, novoHandler(fila))
	if err := http.ListenAndServe(":"+portaOrigem, mux); err != nil {
		log.Fatalf("Erro na porta %s: %v", portaOrigem, err)
	}
}

func main() {
	// Rota 1: 8081 → 3129
	fila1 := make(chan DadosProxy, 10000)
	go processadorFila(fila1, "3129")
	go iniciarServidor("8081", fila1)

	// Rota 2: 8082 → 3130
	fila2 := make(chan DadosProxy, 10000)
	go processadorFila(fila2, "3130")
	go iniciarServidor("8082", fila2)

	// Manter o programa rodando indefinidamente
	select {}
}
