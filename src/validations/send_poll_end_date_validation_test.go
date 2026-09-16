package validations

import (
	"context"
	"testing"
	"time"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolvePollEndMillis locks the conversion from the human friendly
// end_date + end_time pair (Brasília time) into the Unix milliseconds used
// downstream. Fixed references: 16/09/2026 20:35:45 BRT == 23:35:45 UTC ==
// 1789601745000 ms, and 20:35 BRT == 23:35:00 UTC == 1789601700000 ms.
func TestResolvePollEndMillis(t *testing.T) {
	t.Run("numeric end_time passes through unchanged", func(t *testing.T) {
		millis, err := ResolvePollEndMillis(pollEndTimeMillis(1789938000000), nil)
		require.NoError(t, err)
		require.NotNil(t, millis)
		assert.Equal(t, int64(1789938000000), *millis)
	})

	t.Run("end_date (ISO) + end_time resolves to Brasília milliseconds", func(t *testing.T) {
		millis, err := ResolvePollEndMillis(pollEndTimeClock("20:35:45"), strPtr("2026-09-16"))
		require.NoError(t, err)
		require.NotNil(t, millis)
		assert.Equal(t, int64(1789601745000), *millis)
	})

	t.Run("end_date (Brazilian) + end_time without seconds", func(t *testing.T) {
		millis, err := ResolvePollEndMillis(pollEndTimeClock("20:35"), strPtr("16/09/2026"))
		require.NoError(t, err)
		require.NotNil(t, millis)
		assert.Equal(t, int64(1789601700000), *millis)
	})

	t.Run("no fields at all means no deadline", func(t *testing.T) {
		millis, err := ResolvePollEndMillis(nil, nil)
		require.NoError(t, err)
		assert.Nil(t, millis)
	})

	t.Run("clock without end_date is refused", func(t *testing.T) {
		_, err := ResolvePollEndMillis(pollEndTimeClock("20:35:45"), nil)
		assert.Equal(t, pkgError.ValidationError(`end_time "HH:MM:SS" requires end_date`), err)
	})

	t.Run("end_date without clock is refused", func(t *testing.T) {
		_, err := ResolvePollEndMillis(nil, strPtr("2026-09-16"))
		assert.Equal(t, pkgError.ValidationError(`end_date requires end_time "HH:MM:SS"`), err)
	})

	t.Run("end_date cannot be combined with numeric end_time", func(t *testing.T) {
		_, err := ResolvePollEndMillis(pollEndTimeMillis(1789938000000), strPtr("2026-09-16"))
		assert.Equal(t, pkgError.ValidationError(`end_date cannot be combined with a numeric end_time (milliseconds); use end_time "HH:MM:SS" together with end_date, or end_time alone in milliseconds`), err)
	})

	t.Run("invalid date is refused", func(t *testing.T) {
		for _, date := range []string{"2026-02-30", "2026-13-01", "16/13/2026", "16-09-2026", "2026/09/16", "qualquer coisa"} {
			_, err := ResolvePollEndMillis(pollEndTimeClock("20:35:45"), strPtr(date))
			assert.Equal(t, pkgError.ValidationError(`end_date must be "YYYY-MM-DD" or "DD/MM/YYYY"`), err, "date %q", date)
		}
	})

	t.Run("invalid clock is refused", func(t *testing.T) {
		for _, clock := range []string{"25:00", "20:60", "20:35:99", "20h35", "2035", "", "20:35:45:10"} {
			_, err := ResolvePollEndMillis(pollEndTimeClock(clock), strPtr("2026-09-16"))
			assert.Error(t, err, "clock %q must be refused", clock)
		}
	})
}

// TestValidateSendPoll_EndDateAndClock covers the validation rules of the
// human friendly deadline: same future / 1-year window as the milliseconds
// format, but with messages that name the fields the client actually sent.
func TestValidateSendPoll_EndDateAndClock(t *testing.T) {
	baseRequest := func() domainSend.PollRequest {
		return domainSend.PollRequest{
			BaseRequest: domainSend.BaseRequest{Phone: "120363000000000000@g.us"},
			Question:    "Qual o melhor horário para a reunião?",
			Options:     []string{"09h", "14h", "16h"},
			MaxAnswer:   1,
		}
	}

	deadlineIn := func(d time.Duration) (string, string) {
		moment := time.Now().Add(d).In(pollDeadlineLocation)
		return moment.Format("2006-01-02"), moment.Format("15:04:05")
	}

	t.Run("success with date and clock in the future", func(t *testing.T) {
		date, clock := deadlineIn(2 * time.Hour)
		request := baseRequest()
		request.EndDate = strPtr(date)
		request.EndTime = pollEndTimeClock(clock)
		assert.NoError(t, ValidateSendPoll(context.Background(), request))
	})

	t.Run("success with Brazilian date and HH:MM clock", func(t *testing.T) {
		moment := time.Now().Add(2 * time.Hour).In(pollDeadlineLocation)
		request := baseRequest()
		request.EndDate = strPtr(moment.Format("02/01/2006"))
		request.EndTime = pollEndTimeClock(moment.Format("15:04"))
		assert.NoError(t, ValidateSendPoll(context.Background(), request))
	})

	t.Run("error when the deadline is in the past", func(t *testing.T) {
		date, clock := deadlineIn(-1 * time.Minute)
		request := baseRequest()
		request.EndDate = strPtr(date)
		request.EndTime = pollEndTimeClock(clock)
		err := ValidateSendPoll(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be in the future")
	})

	t.Run("error when the deadline is more than one year ahead", func(t *testing.T) {
		date, clock := deadlineIn(366 * 24 * time.Hour)
		request := baseRequest()
		request.EndDate = strPtr(date)
		request.EndTime = pollEndTimeClock(clock)
		err := ValidateSendPoll(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at most 1 year in the future")
	})

	t.Run("error when clock comes without date", func(t *testing.T) {
		request := baseRequest()
		request.EndTime = pollEndTimeClock("20:35:45")
		assert.Equal(t, pkgError.ValidationError(`end_time "HH:MM:SS" requires end_date`),
			ValidateSendPoll(context.Background(), request))
	})

	t.Run("error when date comes without clock", func(t *testing.T) {
		request := baseRequest()
		request.EndDate = strPtr("2026-09-16")
		assert.Equal(t, pkgError.ValidationError(`end_date requires end_time "HH:MM:SS"`),
			ValidateSendPoll(context.Background(), request))
	})

	t.Run("error when mixing date with numeric end_time", func(t *testing.T) {
		request := baseRequest()
		request.EndDate = strPtr("2026-09-16")
		request.EndTime = pollEndTimeMillis(time.Now().Add(2 * time.Hour).UnixMilli())
		assert.Equal(t,
			pkgError.ValidationError(`end_date cannot be combined with a numeric end_time (milliseconds); use end_time "HH:MM:SS" together with end_date, or end_time alone in milliseconds`),
			ValidateSendPoll(context.Background(), request))
	})
}

func strPtr(value string) *string {
	return &value
}

func pollEndTimeClock(value string) *domainSend.PollEndTime {
	return &domainSend.PollEndTime{Clock: &value}
}
