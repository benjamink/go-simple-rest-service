package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/rand"
)

const (
	maxWrong = 6
)

// player struct to seed record player data.
type player struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	CorrectGuesses   int    `json:"correctGuesses"`
	IncorrectGuesses int    `json:"incorrectGuesses"`
	Wins             int    `json:"wins"`
	Losses           int    `json:"losses"`
}

var players = []player{
	{ID: 1, Name: "John Doe", CorrectGuesses: 0, IncorrectGuesses: 0, Wins: 0, Losses: 0},
}

var guessedLetters []string
var currentWord string

// words slice to seed record words data.
var words = []string{
	"toad",
	"balls",
	"unicorn",
	"penguin",
	"squirrel",
}

func main() {
	rand.Seed(uint64(time.Now().UnixNano()))
	selectWord(&currentWord)
	router := gin.New()

	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	router.Use(gin.Recovery())

	router.GET("/players", getPlayers)
	router.GET("/word", func(c *gin.Context) {
		getCurrentWord(c, &currentWord)
	})
	router.GET("/newGame", func(c *gin.Context) {
		resetGame(c, &currentWord)
	})
	router.POST("/player", createPlayer)
	router.GET("/guess/:id/:guess", func(c *gin.Context) {
		guessLetter(c, &currentWord)
	})

	router.Run("localhost:8080")
}

func selectWord(currentWord *string) {
	new := words[rand.Intn(len(words))]
	for new == *currentWord {
		new = words[rand.Intn(len(words))]
	}
	*currentWord = new
}

func resetGame(c *gin.Context, cur *string) {
	selectWord(cur)

	for i, _ := range players {
		players[i].CorrectGuesses = 0
		players[i].IncorrectGuesses = 0
	}

	guessedLetters = []string{}

	c.IndentedJSON(http.StatusOK, "New game started!")
}

func getPlayers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, players)
}

func createPlayer(c *gin.Context) {
	var newPlayer player
	nextIdx := len(players) + 1

	if err := c.BindJSON(&newPlayer); err != nil {
		return
	}

	newPlayer.ID = nextIdx
	players = append(players, newPlayer)
	gin.Logger()
	c.IndentedJSON(http.StatusCreated, newPlayer)
}

func getCurrentWord(c *gin.Context, currentWord *string) {
	c.IndentedJSON(http.StatusOK, currentWord)
}

func updateGuessList(g string) []string {
    if !slices.Contains(guessedLetters, g) {
        guessedLetters = append(guessedLetters, g)
    }
    return guessedLetters
}

func checkIfWinner(correctGuesses int, currentWord *string) bool {
    return correctGuesses >= len(*currentWord)
}

func doTurn(c *gin.Context, player *player, guess string, currentWord *string) {
    if strings.Contains(*currentWord, guess) {
        handleCorrectGuess(c, player, guess, currentWord)
    } else {
        handleIncorrectGuess(c, player, guess)
    }
}

func handleCorrectGuess(c *gin.Context, player *player, guess string, currentWord *string) {
    if !slices.Contains(guessedLetters, guess) {
        player.CorrectGuesses += strings.Count(*currentWord, guess)
    }
    guessed := updateGuessList(guess)
    if checkIfWinner(player.CorrectGuesses, currentWord) {
        respondWithJSON(c, http.StatusOK, player, "You won!", true, guessed)
        resetGame(c, currentWord)
    } else {
        respondWithJSON(c, http.StatusOK, player, "Correct guess!", true, guessed)
    }
}

func handleIncorrectGuess(c *gin.Context, player *player, guess string) {
    player.IncorrectGuesses++
    guessed := updateGuessList(guess)
    respondWithJSON(c, http.StatusOK, player, "Incorrect guess!", false, guessed)
}

func guessLetter(c *gin.Context, currentWord *string) {
    id, _ := strconv.Atoi(c.Param("id"))
    guess := c.Param("guess")

    for i, p := range players {
        if p.ID == id {
            doTurn(c, &players[i], guess, currentWord)
            return
        }
    }
    c.IndentedJSON(http.StatusNotFound, gin.H{"msg": "Player not found"})
}

func respondWithJSON(c *gin.Context, status int, player *player, msg string, isCorrect bool, guessedLetters []string) {
    response := gin.H{
        "player":           player,
        "msg":              msg,
        "isCorrect":        isCorrect,
        "correctGuesses":   player.CorrectGuesses,
        "incorrectGuesses": player.IncorrectGuesses,
        "guessedLetters":   guessedLetters,
    }
    c.IndentedJSON(status, response)
}