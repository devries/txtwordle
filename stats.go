package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/charm/kv"
)

type WordleStats struct {
	CurrentStreak  int            `json:"currentStreak"`
	MaxStreak      int            `json:"maxStreak"`
	Guesses        map[string]int `json:"guesses"`
	WinPercentage  int            `json:"winPercentage"`
	GamesPlayed    int            `json:"gamesPlayed"`
	GamesWon       int            `json:"gamesWon"`
	AverageGuesses int            `json:"averageGuesses"`
}

func decodeStats(data []byte) (WordleStats, error) {
	var ret WordleStats

	err := json.Unmarshal(data, &ret)
	return ret, err
}

func encodeStats(s WordleStats) ([]byte, error) {
	b, err := json.Marshal(s)

	return b, err
}

func getStats() (WordleStats, error) {
	var ret WordleStats

	db, err := kv.OpenWithDefaults("txtwordle-stats")
	if err != nil {
		return ret, err
	}
	defer db.Close()

	db.Sync()

	dat, err := db.Get([]byte("stats"))
	if err != nil {
		return ret, err
	}

	ret, err = decodeStats(dat)
	return ret, err
}

func saveStats(s WordleStats) error {
	dat, err := encodeStats(s)
	if err != nil {
		return err
	}

	db, err := kv.OpenWithDefaults("txtwordle-stats")
	if err != nil {
		return err
	}
	defer db.Close()

	db.Sync()

	err = db.Set([]byte("stats"), dat)
	return err
}

func addWin(s WordleStats, guesses int) WordleStats {
	s.CurrentStreak++
	s.Guesses[strconv.Itoa(guesses)]++
	s.GamesPlayed++
	s.GamesWon++

	if s.CurrentStreak > s.MaxStreak {
		s.MaxStreak = s.CurrentStreak
	}

	s.WinPercentage = 100 * s.GamesWon / s.GamesPlayed

	sumGuess := 0
	for i, v := range []string{"1", "2", "3", "4", "5", "6"} {
		sumGuess = s.Guesses[v] * (i + 1)
	}

	s.AverageGuesses = sumGuess / s.GamesWon

	return s
}

func addLoss(s WordleStats) WordleStats {
	s.CurrentStreak = 0
	s.GamesPlayed++
	s.WinPercentage = 100 * s.GamesWon / s.GamesPlayed
	s.Guesses["fail"]++

	return s
}

func getStatsInfo(s WordleStats) string {
	var bld strings.Builder

	fmt.Fprintf(&bld, "STATISTICS\n")
	fmt.Fprintf(&bld, "Played: %d\n", s.GamesPlayed)
	fmt.Fprintf(&bld, "Win %%: %d\n", s.WinPercentage)
	fmt.Fprintf(&bld, "Current Streak: %d\n", s.CurrentStreak)
	fmt.Fprintf(&bld, "Max Streak: %d\n\n", s.MaxStreak)
	fmt.Fprintf(&bld, "GUESS DISTRIBUTION\n")
	for _, v := range []string{"1", "2", "3", "4", "5", "6"} {
		fmt.Fprintf(&bld, "%s: %d\n", v, s.Guesses[v])
	}
	fmt.Fprintf(&bld, "\n")

	return (&bld).String()
}
