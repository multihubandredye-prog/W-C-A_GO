package validations

import (
	"context"
	"testing"
	"time"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/stretchr/testify/assert"
)

// TestValidateSendPoll_EndTime covers the optional poll end time (auto-close)
// sent as Unix milliseconds, which the API forwards to WhatsApp unchanged.
func TestValidateSendPoll_EndTime(t *testing.T) {
	const group = "120363000000000000@g.us"

	baseRequest := func() domainSend.PollRequest {
		return domainSend.PollRequest{
			BaseRequest: domainSend.BaseRequest{Phone: group},
			Question:    "Qual o melhor horário para a reunião?",
			Options:     []string{"09h", "14h", "16h"},
			MaxAnswer:   1,
		}
	}

	tests := []struct {
		name    string
		endTime *domainSend.PollEndTime
		err     any
	}{
		{
			name:    "should success without end_time (poll stays open forever)",
			endTime: nil,
			err:     nil,
		},
		{
			name:    "should success with a valid deadline in the future",
			endTime: pollEndTimeMillis(time.Now().Add(2 * time.Hour).UnixMilli()),
			err:     nil,
		},
		{
			name:    "should error when end_time is zero",
			endTime: pollEndTimeMillis(0),
			err:     pkgError.ValidationError("end_time must be a positive Unix timestamp in milliseconds"),
		},
		{
			name:    "should error when end_time is negative",
			endTime: pollEndTimeMillis(-1),
			err:     pkgError.ValidationError("end_time must be a positive Unix timestamp in milliseconds"),
		},
		{
			name:    "should error when end_time is already in the past",
			endTime: pollEndTimeMillis(time.Now().Add(-1 * time.Minute).UnixMilli()),
			err:     pkgError.ValidationError("end_time must be in the future"),
		},
		{
			name:    "should error when end_time is more than one year ahead",
			endTime: pollEndTimeMillis(time.Now().Add(366 * 24 * time.Hour).UnixMilli()),
			err:     pkgError.ValidationError("end_time must be at most 1 year in the future"),
		},
		{
			name:    "should error when end_time was sent in seconds instead of milliseconds",
			endTime: pollEndTimeMillis(1789938000),
			err:     pkgError.ValidationError("end_time looks like Unix seconds (1789938000); send it in milliseconds (1789938000000)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := baseRequest()
			request.EndTime = tt.endTime

			err := ValidateSendPoll(context.Background(), request)
			assert.Equal(t, tt.err, err)
		})
	}
}

// TestValidateSendPoll_ExistingBehaviourUnchanged locks the poll fields that
// already shipped, to make sure adding end_time did not change any of them.
func TestValidateSendPoll_ExistingBehaviourUnchanged(t *testing.T) {
	poll := domainSend.PollRequest{
		BaseRequest: domainSend.BaseRequest{Phone: "120363000000000000@g.us"},
		Question:    "Qual o melhor horário para a reunião?",
		Options:     []string{"09h", "14h", "16h"},
		MaxAnswer:   1,
	}

	assert.NoError(t, ValidateSendPoll(context.Background(), poll))
	assert.Nil(t, poll.EndTime, "end_time must stay optional and empty when not informed")
	assert.Nil(t, poll.EndDate, "end_date must stay optional and empty when not informed")

	// Disappearing-message duration keeps its own validation, untouched.
	disappearing := poll
	disappearing.Duration = intPtr(604800)
	assert.NoError(t, ValidateSendPoll(context.Background(), disappearing))

	invalidDuration := poll
	invalidDuration.Duration = intPtr(1234)
	assert.Equal(t,
		pkgError.ValidationError("duration must be one of: 0 (no expiry), 86400 (24h), 604800 (7d), 7776000 (90d)"),
		ValidateSendPoll(context.Background(), invalidDuration),
	)
}

func pollEndTimeMillis(value int64) *domainSend.PollEndTime {
	return &domainSend.PollEndTime{Millis: &value}
}

func intPtr(value int) *int {
	return &value
}
