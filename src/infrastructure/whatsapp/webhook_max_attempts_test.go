package whatsapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
)

func TestSubmitWebhook_NoRetryWithOneAttempt(t *testing.T) {
	original := config.WhatsappWebhookMaxAttempts
	config.WhatsappWebhookMaxAttempts = 1
	defer func() { config.WhatsappWebhookMaxAttempts = original }()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	err := submitWebhook(context.Background(), map[string]any{"Event": "message"}, server.URL)
	if err == nil {
		t.Fatal("expected error when the webhook always answers 500")
	}
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("expected exactly 1 attempt with max_attempts=1, got %d", got)
	}
}

func TestSubmitWebhook_ZeroClampedToOneAttempt(t *testing.T) {
	original := config.WhatsappWebhookMaxAttempts
	config.WhatsappWebhookMaxAttempts = 0
	defer func() { config.WhatsappWebhookMaxAttempts = original }()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_ = submitWebhook(context.Background(), map[string]any{"Event": "message"}, server.URL)
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("expected clamp of 0 to a single attempt, got %d requests", got)
	}
}

func TestSubmitWebhook_RetriesUntilSuccess(t *testing.T) {
	original := config.WhatsappWebhookMaxAttempts
	config.WhatsappWebhookMaxAttempts = 3
	defer func() { config.WhatsappWebhookMaxAttempts = original }()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&requests, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := submitWebhook(context.Background(), map[string]any{"Event": "message"}, server.URL)
	if err != nil {
		t.Fatalf("expected success on attempt 3, got error: %v", err)
	}
	if got := atomic.LoadInt32(&requests); got != 3 {
		t.Fatalf("expected 3 attempts (2 fails + 1 success), got %d", got)
	}
}
