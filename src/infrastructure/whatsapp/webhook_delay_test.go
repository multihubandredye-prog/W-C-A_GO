package whatsapp

import (
	"testing"
	"time"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
)

func TestThrottleWebhookDelivery_SpacesDeliveries(t *testing.T) {
	originalDelay := config.WhatsappWebhookDelayMS
	config.WhatsappWebhookDelayMS = 60
	defer func() { config.WhatsappWebhookDelayMS = originalDelay }()

	url := "https://throttle-spacing.test/hook"
	start := time.Now()
	for i := 0; i < 3; i++ {
		throttleWebhookDelivery(url)
	}
	elapsed := time.Since(start)

	// First delivery goes immediately, the next two wait one interval each.
	if elapsed < 120*time.Millisecond {
		t.Fatalf("expected at least 120ms of spacing across 3 deliveries, got %v", elapsed)
	}
}

func TestThrottleWebhookDelivery_DisabledWhenZero(t *testing.T) {
	originalDelay := config.WhatsappWebhookDelayMS
	config.WhatsappWebhookDelayMS = 0
	defer func() { config.WhatsappWebhookDelayMS = originalDelay }()

	url := "https://throttle-disabled.test/hook"
	start := time.Now()
	for i := 0; i < 3; i++ {
		throttleWebhookDelivery(url)
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("expected no throttling when delay is 0, took %v", elapsed)
	}
}

func TestThrottleWebhookDelivery_IndependentPerURL(t *testing.T) {
	originalDelay := config.WhatsappWebhookDelayMS
	config.WhatsappWebhookDelayMS = 80
	defer func() { config.WhatsappWebhookDelayMS = originalDelay }()

	start := time.Now()
	for i := 0; i < 2; i++ {
		throttleWebhookDelivery("https://throttle-a.test/hook")
		throttleWebhookDelivery("https://throttle-b.test/hook")
	}
	elapsed := time.Since(start)

	// Each URL has its own gate: alternating URLs never wait for the other
	// URL's interval, so 4 deliveries finish well before 3 full intervals.
	if elapsed >= 240*time.Millisecond {
		t.Fatalf("expected per-URL limiters to work independently, took %v", elapsed)
	}
}
