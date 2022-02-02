package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"
)

func main() {
	r, c := initialize()
	clear()
	hideCursor()

	drawBoard(r, c)

	input := make(chan rune, 1)
	go readKeys(input)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGWINCH)

	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	firstDay := time.Date(2021, time.June, 20, 0, 0, 0, 0, zone)
	now := time.Now()
	days := int(now.Sub(firstDay).Truncate(24*time.Hour) / time.Hour / 24)

	word := strings.ToUpper(wordList[days])
	state := State{[]string{}, []rune{}}

	// Game loop
gameloop:
	for {
		drawState(r, c, state, word)
		if len(state.Guesses) == 6 {
			break gameloop
		}

		select {
		case s := <-sigChan:
			if s == syscall.SIGWINCH {
				r, c = resize()
				clear()
				drawBoard(r, c)
				// Draw current state
			} else {
				break gameloop
			}
		case letter := <-input:
			// Add or delete letter
			move(r-1, 0)
			// fmt.Printf("CHAR: %4d", c)
			switch letter {
			case 10: // Enter
				if len(state.Current) == 5 {
					// Check for win
					guess := string(state.Current)
					state.Guesses = append(state.Guesses, guess)
					state.Current = []rune{}
					if guess == word {
						drawState(r, c, state, word)
						break gameloop
					}
				}
			case 127:
				if len(state.Current) > 0 {
					state.Current = state.Current[:len(state.Current)-1]
				}
			default:
				// Add a letter
				if len(state.Current) < 5 {
					state.Current = append(state.Current, letter)
				}
			}
		}
	}

	move(r-1, 0)
	cleanup()
	res := getCopyPaste(state, word, days)
	fmt.Print(res)
}

func drawBoard(rows, columns int) {
	top := rows/2 - 7
	left := columns/2 - 6

	move(top, left)
	fmt.Printf("┏━┳━┳━┳━┳━┓")
	for i := 0; i < 6; i++ {
		move(top+i*2+1, left)
		fmt.Printf("┃ ┃ ┃ ┃ ┃ ┃")
		move(top+i*2+2, left)
		fmt.Printf("┣━╋━╋━╋━╋━┫")
	}
	move(top+12, left)
	fmt.Printf("┗━┻━┻━┻━┻━┛")

	kbtop := rows/2 + 7
	kbleft := columns/2 - 10

	for k, v := range keyboardPositions {
		move(kbtop+v.Y, kbleft+v.X)
		fmt.Printf("%c", k)
	}
}

func drawState(rows, columns int, state State, word string) {
	correctLetters := []rune(word)

	for y, guess := range state.Guesses {
		for x, letter := range guess {
			// Check if this letter is in the word
			inword := false
			for _, l := range correctLetters {
				if letter == l {
					inword = true
					break
				}
			}

			if letter == correctLetters[x] {
				setGreen(rows, columns, letter)
				greenBackground()
			} else if inword {
				setYellow(rows, columns, letter)
				yellowBackground()
			} else {
				setGray(rows, columns, letter)
				grayBackground()
			}

			drawLetter(rows, columns, x, y, letter)
			defaultBackground()
		}
	}

	currentRow := len(state.Guesses)
	for i := 0; i < 5; i++ {
		if i < len(state.Current) {
			drawLetter(rows, columns, i, currentRow, state.Current[i])
		} else {
			drawLetter(rows, columns, i, currentRow, ' ')
		}
	}
}

func getCopyPaste(state State, word string, days int) string {
	correctLetters := []rune(word)

	var bld strings.Builder

	fmt.Fprintf(&bld, "Wordle %d %d/6\n\n", days+1, len(state.Guesses))
	for _, guess := range state.Guesses {
		for x, letter := range guess {
			// Check if this letter is in the word
			inword := false
			for _, l := range correctLetters {
				if letter == l {
					inword = true
					break
				}
			}

			if letter == correctLetters[x] {
				fmt.Fprintf(&bld, "🟩")
			} else if inword {
				fmt.Fprintf(&bld, "🟨")
			} else {
				fmt.Fprintf(&bld, "⬜")
			}
		}
		fmt.Fprintf(&bld, "\n")
	}

	return (&bld).String()
}

func drawLetter(rows, columns int, x, y int, letter rune) {
	x0 := columns/2 - 5
	y0 := rows/2 - 6

	colPos := x0 + 2*x
	rowPos := y0 + 2*y

	move(rowPos, colPos)
	fmt.Printf("%c", letter)
}

func greenBackground() {
	fmt.Print("\x1b[97;42m")
}

func grayBackground() {
	fmt.Print("\x1b[97;100m")
}

func yellowBackground() {
	fmt.Print("\x1b[97;43m")
}

func defaultBackground() {
	fmt.Print("\x1b[39;49m")
}

func setGreen(rows, columns int, letter rune) {
	kbtop := rows/2 + 7
	kbleft := columns/2 - 10

	pt := keyboardPositions[letter]

	greenBackground()
	move(kbtop+pt.Y, kbleft+pt.X)
	fmt.Printf("%c", letter)
	defaultBackground()
}

func setYellow(rows, columns int, letter rune) {
	kbtop := rows/2 + 7
	kbleft := columns/2 - 10

	pt := keyboardPositions[letter]

	yellowBackground()
	move(kbtop+pt.Y, kbleft+pt.X)
	fmt.Printf("%c", letter)
	defaultBackground()
}

func setGray(rows, columns int, letter rune) {
	kbtop := rows/2 + 7
	kbleft := columns/2 - 10

	pt := keyboardPositions[letter]

	grayBackground()
	move(kbtop+pt.Y, kbleft+pt.X)
	fmt.Printf("%c", letter)
	defaultBackground()
}

type Point struct {
	X int
	Y int
}

var keyboardPositions = map[rune]Point{
	'Q': {0, 0},
	'W': {2, 0},
	'E': {4, 0},
	'R': {6, 0},
	'T': {8, 0},
	'Y': {10, 0},
	'U': {12, 0},
	'I': {14, 0},
	'O': {16, 0},
	'P': {18, 0},
	'A': {1, 1},
	'S': {3, 1},
	'D': {5, 1},
	'F': {7, 1},
	'G': {9, 1},
	'H': {11, 1},
	'J': {13, 1},
	'K': {15, 1},
	'L': {17, 1},
	'Z': {3, 2},
	'X': {5, 2},
	'C': {7, 2},
	'V': {9, 2},
	'B': {11, 2},
	'N': {13, 2},
	'M': {15, 2},
}

var wordList = []string{"rebut", "sissy", "humph", "awake", "blush", "focal", "evade", "naval", "serve", "heath", "dwarf", "model", "karma", "stink", "grade", "quiet", "bench", "abate", "feign", "major", "death", "fresh", "crust", "stool", "colon", "abase", "marry", "react", "batty", "pride", "floss", "helix", "croak", "staff", "paper", "unfed", "whelp", "trawl", "outdo", "adobe", "crazy", "sower", "repay", "digit", "crate", "cluck", "spike", "mimic", "pound", "maxim", "linen", "unmet", "flesh", "booby", "forth", "first", "stand", "belly", "ivory", "seedy", "print", "yearn", "drain", "bribe", "stout", "panel", "crass", "flume", "offal", "agree", "error", "swirl", "argue", "bleed", "delta", "flick", "totem", "wooer", "front", "shrub", "parry", "biome", "lapel", "start", "greet", "goner", "golem", "lusty", "loopy", "round", "audit", "lying", "gamma", "labor", "islet", "civic", "forge", "corny", "moult", "basic", "salad", "agate", "spicy", "spray", "essay", "fjord", "spend", "kebab", "guild", "aback", "motor", "alone", "hatch", "hyper", "thumb", "dowry", "ought", "belch", "dutch", "pilot", "tweed", "comet", "jaunt", "enema", "steed", "abyss", "growl", "fling", "dozen", "boozy", "erode", "world", "gouge", "click", "briar", "great", "altar", "pulpy", "blurt", "coast", "duchy", "groin", "fixer", "group", "rogue", "badly", "smart", "pithy", "gaudy", "chill", "heron", "vodka", "finer", "surer", "radio", "rouge", "perch", "retch", "wrote", "clock", "tilde", "store", "prove", "bring", "solve", "cheat", "grime", "exult", "usher", "epoch", "triad", "break", "rhino", "viral", "conic", "masse", "sonic", "vital", "trace", "using", "peach", "champ", "baton", "brake", "pluck", "craze", "gripe", "weary", "picky", "acute", "ferry", "aside", "tapir", "troll", "unify", "rebus", "boost", "truss", "siege", "tiger", "banal", "slump", "crank", "gorge", "query", "drink", "favor", "abbey", "tangy", "panic", "solar", "shire", "proxy", "point", "robot", "prick", "wince", "crimp", "knoll", "sugar", "whack", "mount", "perky", "could", "wrung", "light", "those", "moist", "shard", "pleat", "aloft", "skill", "elder", "frame", "humor", "pause", "ulcer", "ultra", "robin", "cynic", "agora", "aroma", "caulk", "shake", "pupal", "dodge", "swill", "tacit", "other", "thorn", "trove", "bloke", "vivid", "spill", "chant", "choke", "rupee", "nasty", "mourn", "ahead", "brine", "cloth", "hoard", "sweet", "month", "lapse", "watch", "today", "focus", "smelt", "tease", "cater", "movie", "lynch", "saute", "allow", "renew", "their", "slosh", "purge", "chest", "depot", "epoxy", "nymph", "found", "shall", "harry", "stove", "lowly", "snout", "trope", "fewer", "shawl", "natal", "fibre", "comma", "foray", "scare", "stair", "black", "squad", "royal", "chunk", "mince", "slave", "shame", "cheek", "ample", "flair", "foyer", "cargo", "oxide", "plant", "olive", "inert", "askew", "heist", "shown", "zesty", "hasty", "trash", "fella", "larva", "forgo", "story", "hairy", "train", "homer", "badge", "midst", "canny", "fetus", "butch", "farce", "slung", "tipsy", "metal", "yield", "delve", "being", "scour", "glass", "gamer", "scrap", "money", "hinge", "album", "vouch", "asset", "tiara", "crept", "bayou", "atoll", "manor", "creak", "showy", "phase", "froth", "depth", "gloom", "flood", "trait", "girth", "piety", "payer", "goose", "float", "donor", "atone", "primo", "apron", "blown", "cacao", "loser", "input", "gloat", "awful", "brink", "smite", "beady", "rusty", "retro", "droll", "gawky", "hutch", "pinto", "gaily", "egret", "lilac", "sever", "field", "fluff", "hydro", "flack", "agape", "wench", "voice", "stead", "stalk", "berth", "madam", "night", "bland", "liver", "wedge", "augur", "roomy", "wacky", "flock", "angry", "bobby", "trite", "aphid", "tryst", "midge", "power", "elope", "cinch", "motto", "stomp", "upset", "bluff", "cramp", "quart", "coyly", "youth", "rhyme", "buggy", "alien", "smear", "unfit", "patty", "cling", "glean", "label", "hunky", "khaki", "poker", "gruel", "twice", "twang", "shrug", "treat", "unlit", "waste", "merit", "woven", "octal", "needy", "clown", "widow", "irony", "ruder", "gauze", "chief", "onset", "prize", "fungi", "charm", "gully", "inter", "whoop", "taunt", "leery", "class", "theme", "lofty", "tibia", "booze", "alpha", "thyme", "eclat", "doubt", "parer", "chute", "stick", "trice", "alike", "sooth", "recap", "saint", "liege", "glory", "grate", "admit", "brisk", "soggy", "usurp", "scald", "scorn", "leave", "twine", "sting", "bough", "marsh", "sloth", "dandy", "vigor", "howdy", "enjoy"}

func readKeys(input chan rune) {
	reader := bufio.NewReader(os.Stdin)

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			panic(err)
		}
		input <- unicode.ToUpper(char)
	}
}

type State struct {
	Guesses []string
	Current []rune
}
