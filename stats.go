package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
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

func NewWordleStats() WordleStats {
	return WordleStats{
		CurrentStreak:  0,
		MaxStreak:      0,
		Guesses:        map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0, "6": 0, "fail": 0},
		WinPercentage:  0,
		GamesPlayed:    0,
		GamesWon:       0,
		AverageGuesses: 0,
	}
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

func getFileStats() (WordleStats, error) {
	stats := NewWordleStats()
	configFile, err := getConfigFilename()
	if err != nil {
		return stats, fmt.Errorf("unable to get stats: %s", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return stats, fmt.Errorf("unable to get stats: %s", err)
	}

	stats, err = decodeStats(data)
	return stats, err
}

func getConfigFilename() (string, error) {
	var configFile string

	confdir, err := os.UserConfigDir()
	if err != nil {
		confdir = "."
	}

	appconfdir := path.Join(confdir, "txtwordle")
	err = os.MkdirAll(appconfdir, 0750)
	if err != nil {
		return configFile, fmt.Errorf("unable to create config directory %s", appconfdir)
	}

	return path.Join(appconfdir, "config.json"), nil
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

func saveFileStats(s WordleStats) error {
	dat, err := encodeStats(s)
	if err != nil {
		return err
	}

	configFile, err := getConfigFilename()
	if err != nil {
		return fmt.Errorf("unable to write stats: %s", err)
	}

	if err := os.WriteFile(configFile, dat, 0640); err != nil {
		return fmt.Errorf("unable to write stats: %s", err)
	}

	return nil
}

func addWin(s WordleStats, guesses int) WordleStats {
	s.CurrentStreak++
	s.Guesses[strconv.Itoa(guesses)]++
	s.GamesPlayed++
	s.GamesWon++

	if s.CurrentStreak > s.MaxStreak {
		s.MaxStreak = s.CurrentStreak
	}

	s.WinPercentage = int(math.Round(100.0 * float64(s.GamesWon) / float64(s.GamesPlayed)))

	sumGuess := 0
	for i, v := range []string{"1", "2", "3", "4", "5", "6"} {
		sumGuess = s.Guesses[v] * (i + 1)
	}

	s.AverageGuesses = int(math.Round(float64(sumGuess) / float64(s.GamesWon)))

	return s
}

func addLoss(s WordleStats) WordleStats {
	s.CurrentStreak = 0
	s.GamesPlayed++
	s.WinPercentage = int(math.Round(100.0 * float64(s.GamesWon) / float64(s.GamesPlayed)))
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
	var maxDist int
	for _, v := range []string{"1", "2", "3", "4", "5", "6"} {
		if s.Guesses[v] > maxDist {
			maxDist = s.Guesses[v]
		}
	}

	for _, v := range []string{"1", "2", "3", "4", "5", "6"} {
		fraction := float64(s.Guesses[v]) / float64(maxDist)
		fmt.Fprintf(&bld, "%s: %s %d\n", v, getHorizontalBar(40, fraction), s.Guesses[v])
	}
	fmt.Fprintf(&bld, "\n")

	return (&bld).String()
}

func getHorizontalBar(maxLength int, fraction float64) string {
	maxEighths := float64(maxLength * 8)
	eighths := int(maxEighths * fraction)

	wholeBlocks := eighths / 8
	remainder := eighths % 8

	var bld strings.Builder

	for i := 0; i < wholeBlocks; i++ {
		fmt.Fprint(&bld, "\u2588")
	}
	switch remainder {
	case 1:
		fmt.Fprint(&bld, "\u258f")
	case 2:
		fmt.Fprint(&bld, "\u258e")
	case 3:
		fmt.Fprint(&bld, "\u258d")
	case 4:
		fmt.Fprint(&bld, "\u258c")
	case 5:
		fmt.Fprint(&bld, "\u258b")
	case 6:
		fmt.Fprint(&bld, "\u258a")
	case 7:
		fmt.Fprint(&bld, "\u2589")
	}

	return (&bld).String()
}
