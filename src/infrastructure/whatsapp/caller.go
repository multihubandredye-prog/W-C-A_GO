package whatsapp

import (
	"encoding/json"
	"io"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"

	"github.com/purpshell/meowcaller"
)

// NewMeowCaller builds the meowcaller Client for a whatsmeow client.
//
// CRITICAL: it must be constructed BEFORE the whatsmeow client connects.
// meowcaller installs a low-level <ack>/<call> interception at construction
// time (engine.installCallAckHook refuses to run on an already connected
// client). That interception is how the engine learns the relay endpoint when
// the peer answers — without it the offer still goes out (the phone rings) but
// media never starts and the callee is stuck on "Conectando..." forever.
//
// The engine's zerolog output is bridged into the API's logrus logs, so call
// lifecycle events ("first RTP sent to relay", "first RTP decoded from
// relay", relay errors, "raw call adapter is unavailable") are visible when
// diagnosing calls.
func NewMeowCaller(wa *whatsmeow.Client) *meowcaller.Client {
	if wa == nil {
		return nil
	}

	level := zerolog.InfoLevel
	if config.AppDebug {
		level = zerolog.DebugLevel
	}

	logger := zerolog.New(meowcallerLogBridge{}).Level(level).With().Timestamp().Logger()
	return meowcaller.NewClient(wa, meowcaller.WithLogger(logger))
}

// meowcallerLogBridge translates meowcaller's zerolog JSON lines into logrus
// entries prefixed with [meowcaller].
type meowcallerLogBridge struct{}

func (meowcallerLogBridge) Write(p []byte) (int, error) {
	var event struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(p, &event); err != nil {
		// Not a JSON line (should not happen with zerolog): forward raw.
		logrus.Infof("[meowcaller] %s", string(p))
		return len(p), nil
	}

	message := "[meowcaller] " + event.Message
	if event.Error != "" {
		message += ": " + event.Error
	}

	switch event.Level {
	case "trace", "debug":
		logrus.Debug(message)
	case "warn":
		logrus.Warn(message)
	case "error", "fatal", "panic":
		logrus.Error(message)
	default: // info and anything unexpected
		logrus.Info(message)
	}
	return len(p), nil
}

// ensure the bridge satisfies io.Writer at compile time.
var _ io.Writer = meowcallerLogBridge{}
