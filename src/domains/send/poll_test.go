package send

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPollRequestJSONEndTimeNumeric locks the original JSON contract: end_time
// as a number is Unix milliseconds. The REST handler feeds the request body
// straight into this struct with Fiber's BodyParser, so the tag name is what
// the endpoint publishes.
func TestPollRequestJSONEndTimeNumeric(t *testing.T) {
	body := []byte(`{
		"phone": "120363000000000000@g.us",
		"question": "Qual o melhor horário para a reunião?",
		"options": ["09h", "14h", "16h"],
		"max_answer": 1,
		"end_time": 1789938000000
	}`)

	var request PollRequest
	require.NoError(t, json.Unmarshal(body, &request))

	require.NotNil(t, request.EndTime)
	require.NotNil(t, request.EndTime.Millis)
	assert.Equal(t, int64(1789938000000), *request.EndTime.Millis)
	assert.Nil(t, request.EndTime.Clock)
	assert.Nil(t, request.EndDate)
	assert.Equal(t, "120363000000000000@g.us", request.Phone)
	assert.Equal(t, []string{"09h", "14h", "16h"}, request.Options)
	assert.Equal(t, 1, request.MaxAnswer)

	// Round-trip: number in, number out.
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"end_time":1789938000000`)
}

// TestPollRequestJSONEndTimeClock locks the human friendly format: end_time
// as a time-of-day string, always paired with end_date.
func TestPollRequestJSONEndTimeClock(t *testing.T) {
	body := []byte(`{
		"phone": "120363000000000000@g.us",
		"question": "Qual o melhor horário para a reunião?",
		"options": ["09h", "14h", "16h"],
		"max_answer": 1,
		"end_date": "2026-09-16",
		"end_time": "20:35:45"
	}`)

	var request PollRequest
	require.NoError(t, json.Unmarshal(body, &request))

	require.NotNil(t, request.EndTime)
	require.NotNil(t, request.EndTime.Clock)
	assert.Equal(t, "20:35:45", *request.EndTime.Clock)
	assert.Nil(t, request.EndTime.Millis)
	require.NotNil(t, request.EndDate)
	assert.Equal(t, "2026-09-16", *request.EndDate)

	// Round-trip: string in, string out.
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"end_time":"20:35:45"`)
	assert.Contains(t, string(encoded), `"end_date":"2026-09-16"`)
}

// TestPollRequestJSONEndTimeNullAndInvalid: JSON null keeps the field unset,
// and unsupported types are a parse error instead of being silently dropped.
func TestPollRequestJSONEndTimeNullAndInvalid(t *testing.T) {
	var nullRequest PollRequest
	require.NoError(t, json.Unmarshal([]byte(`{"end_time": null}`), &nullRequest))
	assert.Nil(t, nullRequest.EndTime)

	var boolValue PollRequest
	assert.Error(t, json.Unmarshal([]byte(`{"end_time": true}`), &boolValue))

	var listValue PollRequest
	assert.Error(t, json.Unmarshal([]byte(`{"end_time": ["20:35"]}`), &listValue))
}

// TestPollEndTimeUnmarshalText covers the form-data path: Fiber's form parser
// sets custom types through encoding.TextUnmarshaler, where everything is a
// string — a purely numeric value means milliseconds, anything else is a
// time-of-day.
func TestPollEndTimeUnmarshalText(t *testing.T) {
	var numeric PollEndTime
	require.NoError(t, numeric.UnmarshalText([]byte("1789938000000")))
	require.NotNil(t, numeric.Millis)
	assert.Equal(t, int64(1789938000000), *numeric.Millis)
	assert.Nil(t, numeric.Clock)

	var clock PollEndTime
	require.NoError(t, clock.UnmarshalText([]byte("20:35:45")))
	require.NotNil(t, clock.Clock)
	assert.Equal(t, "20:35:45", *clock.Clock)
	assert.Nil(t, clock.Millis)

	var blank PollEndTime
	require.NoError(t, blank.UnmarshalText([]byte(" ")))
	assert.Nil(t, blank.Millis)
	assert.Nil(t, blank.Clock)
}

// TestPollRequestJSONWithoutEndTime keeps the previous payload working exactly as
// before: no end_time means no deadline.
func TestPollRequestJSONWithoutEndTime(t *testing.T) {
	body := []byte(`{
		"phone": "120363000000000000@g.us",
		"question": "Qual o melhor horário para a reunião?",
		"options": ["09h", "14h", "16h"],
		"max_answer": 1
	}`)

	var request PollRequest
	require.NoError(t, json.Unmarshal(body, &request))

	assert.Nil(t, request.EndTime)
	assert.Nil(t, request.EndDate)
	assert.Equal(t, "Qual o melhor horário para a reunião?", request.Question)

	// end_time is omitempty, so it must not leak into outgoing payloads either.
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "end_time")
	assert.NotContains(t, string(encoded), "end_date")
}
