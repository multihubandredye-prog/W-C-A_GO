package whatsapp

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

// AutoCallPolicy is the per-device rule. This is a policy engine only; it does
// not read any file, update WhatsApp state, or enable answering by itself.
type AutoCallPolicy struct {
	Mode            string   `json:"mode"`    // all, allow, ignore
	Numbers         []string `json:"numbers"` // canonical international digits
	ScheduleEnabled bool     `json:"schedule_enabled"`
	Days            []int    `json:"days"`     // Sunday=0 through Saturday=6
	Start           string   `json:"start"`    // HH:MM, inclusive
	End             string   `json:"end"`      // HH:MM, exclusive; before start means overnight
	Timezone        string   `json:"timezone"` // IANA name
}

var internationalDigits = regexp.MustCompile(`^[1-9][0-9]{6,14}$`)

func validPolicyClock(s string) (int, error) {
	if len(s) != 5 || s[2] != ':' || s[0] < '0' || s[0] > '9' || s[1] < '0' || s[1] > '9' || s[3] < '0' || s[3] > '9' || s[4] < '0' || s[4] > '9' {
		return 0, errors.New("time must use HH:MM")
	}
	h, e1 := strconv.Atoi(s[:2])
	m, e2 := strconv.Atoi(s[3:])
	if e1 != nil || e2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, errors.New("invalid time")
	}
	return h*60 + m, nil
}
func (p AutoCallPolicy) Validate() error {
	if p.Mode != "all" && p.Mode != "allow" && p.Mode != "ignore" {
		return errors.New("invalid mode")
	}
	if len(p.Numbers) > 5000 {
		return errors.New("too many numbers")
	}
	seen := map[string]bool{}
	for _, n := range p.Numbers {
		if !internationalDigits.MatchString(n) {
			return fmt.Errorf("invalid phone number")
		}
		if seen[n] {
			return errors.New("duplicate phone number")
		}
		seen[n] = true
	}
	if p.ScheduleEnabled {
		if p.Timezone == "" || p.Timezone == "Local" {
			return errors.New("explicit IANA timezone required")
		}
		if _, err := time.LoadLocation(p.Timezone); err != nil {
			return errors.New("invalid timezone")
		}
		start, err := validPolicyClock(p.Start)
		if err != nil {
			return err
		}
		end, err := validPolicyClock(p.End)
		if err != nil {
			return err
		}
		if start == end {
			return errors.New("start and end must differ")
		}
		if len(p.Days) == 0 || len(p.Days) > 7 {
			return errors.New("select at least one day")
		}
		days := map[int]bool{}
		for _, day := range p.Days {
			if day < 0 || day > 6 || days[day] {
				return errors.New("invalid or duplicate day")
			}
			days[day] = true
		}
	}
	return nil
}

// Allows returns false for unresolved phone numbers in filtered modes, even
// when ignoring numbers, to avoid answering someone whose identity is unknown.
func (p AutoCallPolicy) Allows(number string, now time.Time) (bool, error) {
	if err := p.Validate(); err != nil {
		return false, err
	}
	if p.Mode != "all" {
		if !internationalDigits.MatchString(number) {
			return false, nil
		}
		selected := false
		for _, n := range p.Numbers {
			if n == number {
				selected = true
				break
			}
		}
		if p.Mode == "allow" && !selected || p.Mode == "ignore" && selected {
			return false, nil
		}
	}
	if !p.ScheduleEnabled {
		return true, nil
	}
	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		return false, err
	}
	t := now.In(loc)
	cur := t.Hour()*60 + t.Minute()
	start, _ := validPolicyClock(p.Start)
	end, _ := validPolicyClock(p.End)
	hasDay := func(day int) bool {
		for _, v := range p.Days {
			if v == day {
				return true
			}
		}
		return false
	}
	day := int(t.Weekday())
	if start < end {
		return hasDay(day) && cur >= start && cur < end, nil
	}
	if cur >= start {
		return hasDay(day), nil
	}
	if cur < end {
		return hasDay((day + 6) % 7), nil
	}
	return false, nil
}

// canonicalManualPhone accepts digits and common presentation punctuation;
// it never guesses country code or resolves a WhatsApp LID.
func canonicalManualPhone(input string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(input) {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '+' || r == ' ' || r == '-' || r == '(' || r == ')':
		default:
			return "", errors.New("invalid phone number")
		}
	}
	n := b.String()
	if !internationalDigits.MatchString(n) {
		return "", errors.New("number must include country code")
	}
	return n, nil
}
