package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// buildTestPollMessage mimics what client.BuildPollCreation returns, without
// needing a connected WhatsApp client.
func buildTestPollMessage() *waE2E.Message {
	return &waE2E.Message{
		PollCreationMessage: &waE2E.PollCreationMessage{
			Name:    proto.String("Qual o melhor horário?"),
			Options: []*waE2E.PollCreationMessage_Option{{OptionName: proto.String("09h")}},
		},
	}
}

// TestApplyPollEndTime_ForwardsMillisecondsUnchanged is the core guarantee of
// the feature: the caller sends milliseconds (Date.now()) and WhatsApp's
// PollCreationMessage.endTime is itself milliseconds, so the wire value must be
// identical. Dividing by 1000 makes the poll arrive already closed (the app
// reads the small value as ms in 1970) — the regression this test locks out.
func TestApplyPollEndTime_ForwardsMillisecondsUnchanged(t *testing.T) {
	// 20/09/2026 21:00:00 UTC
	endTimeMilliseconds := int64(1789938000000)

	msg := buildTestPollMessage()
	applyPollEndTime(msg, &endTimeMilliseconds)

	require.NotNil(t, msg.PollCreationMessage.EndTime)
	assert.Equal(t, int64(1789938000000), msg.PollCreationMessage.GetEndTime(),
		"wire value must be Unix milliseconds, unchanged")

	// The proto must really carry the field (non-zero, encoded).
	encoded, err := proto.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded := &waE2E.Message{}
	require.NoError(t, proto.Unmarshal(encoded, decoded))
	assert.Equal(t, int64(1789938000000), decoded.GetPollCreationMessage().GetEndTime())
	assert.Equal(t, time.Date(2026, time.September, 20, 21, 0, 0, 0, time.UTC),
		time.UnixMilli(decoded.GetPollCreationMessage().GetEndTime()).UTC())
}

// TestApplyPollEndTime_WithoutEndTimeKeepsMessageUntouched guarantees the old
// behaviour: no end_time in, no endTime on the wire.
func TestApplyPollEndTime_WithoutEndTimeKeepsMessageUntouched(t *testing.T) {
	for _, tt := range []struct {
		name    string
		endTime *int64
	}{
		{name: "nil", endTime: nil},
		{name: "zero", endTime: proto.Int64(0)},
		{name: "negative", endTime: proto.Int64(-1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			msg := buildTestPollMessage()
			applyPollEndTime(msg, tt.endTime)

			assert.Nil(t, msg.PollCreationMessage.EndTime, "endTime must stay unset")
			assert.Equal(t, "Qual o melhor horário?", msg.PollCreationMessage.GetName())
			assert.Len(t, msg.PollCreationMessage.GetOptions(), 1)
		})
	}
}

// TestApplyPollEndTime_ToleratesNilMessage makes sure a nil message can't panic.
func TestApplyPollEndTime_ToleratesNilMessage(t *testing.T) {
	endTime := int64(1789938000000)
	assert.NotPanics(t, func() { applyPollEndTime(nil, &endTime) })

	msg := &waE2E.Message{}
	applyPollEndTime(msg, &endTime)
	require.NotNil(t, msg.PollCreationMessage)
	assert.Equal(t, int64(1789938000000), msg.PollCreationMessage.GetEndTime())
}
