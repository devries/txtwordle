package main

import (
	"testing"
)

var tstval = `{"currentStreak":6,"maxStreak":46,"guesses":{"1":0,"2":4,"3":28,"4":26,"5":13,"6":4,"fail":1},"winPercentage":99,"gamesPlayed":76,"gamesWon":75,"averageGuesses":4}`

func TestStorage(t *testing.T) {
	s, err := decodeStats([]byte(tstval))
	if err != nil {
		t.Errorf("Error unmarshaling test value: %s", err)
	}

	err = saveStats(s)
	if err != nil {
		t.Errorf("Error saving stats: %s", err)
	}

	s2, err := getStats()
	if err != nil {
		t.Errorf("Error retrieving stats: %s", err)
	}

	if s.MaxStreak != s2.MaxStreak {
		t.Errorf("Retrieved value does not match sent value\nSent: %+v\nRetrieved: %+v\n", s, s2)
	}
}
