package send

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPollRequestJSONEndTime locks the JSON contract of the poll auto-close
// field. The REST handler feeds the request body straight into this struct with
// Fiber's BodyParser, so the tag name is what the endpoint publishes.
func TestPollRequestJSONEndTime(t *testing.T) {
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
	assert.Equal(t, int64(1789938000000), *request.EndTime)
	assert.Equal(t, "120363000000000000@g.us", request.Phone)
	assert.Equal(t, []string{"09h", "14h", "16h"}, request.Options)
	assert.Equal(t, 1, request.MaxAnswer)
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
	assert.Equal(t, "Qual o melhor horário para a reunião?", request.Question)

	// end_time is omitempty, so it must not leak into outgoing payloads either.
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "end_time")
}
