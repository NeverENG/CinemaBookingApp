package service

import (
	"encoding/json"
	"strings"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

func parsePriceRules(raw string) (map[string]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]int64{}, nil
	}
	var rules map[string]int64
	if err := json.Unmarshal([]byte(raw), &rules); err != nil || rules == nil {
		return nil, domain.ErrSessionInvalid
	}
	for seatType, price := range rules {
		if strings.TrimSpace(seatType) == "" || price <= 0 {
			return nil, domain.ErrSessionInvalid
		}
	}
	return rules, nil
}

func normalizePriceRulesJSON(raw string) (string, error) {
	rules, err := parsePriceRules(raw)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(rules)
	if err != nil {
		return "", domain.ErrSessionInvalid
	}
	return string(encoded), nil
}

func priceForSeat(session *domain.ShowSession, rules map[string]int64, seat domain.Seat) (int64, error) {
	if session.BasePriceCents <= 0 {
		return 0, domain.ErrSessionInvalid
	}
	if price, ok := rules[strings.TrimSpace(seat.Type)]; ok {
		return price, nil
	}
	return session.BasePriceCents, nil
}
