package passgen

import (
	cryptorand "crypto/rand"
	"errors"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type PasswordGenerator struct {
	wordlist []string
	values   Values
}

type Password struct {
	Xkcd     string
	Original string
	Length   int
}

type Values struct {
	MIN_PASSWORD_LENGTH  int               `yaml:"MIN_PASSWORD_LENGTH"`
	MAX_PASSWORD_LENGTH  int               `yaml:"MAX_PASSWORD_LENGTH"`
	BETWEEN_SYMBOLS      string            `yaml:"BETWEEN_SYMBOLS"`
	INSIDE_SYMBOLS       string            `yaml:"INSIDE_SYMBOLS"`
	PASSWORD_PER_ROUTINE int               `yaml:"PASSWORD_PER_ROUTINE"`
	SYMBOL_MAPPING       map[string]string `yaml:"SYMBOL_MAPPING"`
	WORDLIST_PATH        string            `yaml:"WORDLIST_PATH"`
}

var (
	once     sync.Once
	wordList []string
)

// NewPasswordGenerator creates a new PasswordGenerator instance.
func NewPasswordGenerator(values Values) (*PasswordGenerator, error) {
	var err error
	once.Do(func() {
		if strings.HasPrefix(values.WORDLIST_PATH, "http://") || strings.HasPrefix(values.WORDLIST_PATH, "https://") {
			wordList, err = loadWordsFromURL(values.WORDLIST_PATH)
		} else {
			wordList, err = loadWordsFromFile(values.WORDLIST_PATH)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("Error loading wordlist")
		}
	})
	generator := &PasswordGenerator{
		wordlist: wordList,
		values:   values,
	}

	return generator, nil
}

func loadWordsFromURL(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error fetching wordlist from URL: %s", url)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error reading wordlist from URL: %s", url)
		return nil, err
	}
	log.Info().Msgf("Loaded %d words from URL", len(strings.Split(string(data), "\n")))
	return strings.Split(string(data), "\n"), nil
}

func loadWordsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error reading wordlist from file: %s", filename)
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func getSymbol(symbols string) (string, error) {
	// Handle the case where symbols is empty.
	if symbols == "" {
		return "", errors.New("symbol list is empty")
	}
	// Handle the case where symbols has only one char.
	if len(symbols) == 1 {
		return symbols, nil
	}
	symbolIndex := rand.Intn(len(symbols))
	symbol := string(symbols[symbolIndex])

	return symbol, nil
}

func (pg *PasswordGenerator) generator(minPasswordLength int, maxPasswordLength int) (string, string) {
	var symbol string
	passwordWords := make([]string, 3)
	totalLength := 0
	for totalLength < minPasswordLength || totalLength > maxPasswordLength-2 {
		totalLength = 0
		for i := 0; i < 3; i++ {
			randomIndex, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(int64((len(pg.wordlist)))))
			passwordWords[i] = pg.wordlist[randomIndex.Int64()]
			totalLength += len(passwordWords[i])
		}
	}
	plainPassword := strings.Join(passwordWords, " ")
	symbol, err := getSymbol(pg.values.BETWEEN_SYMBOLS)
	if err != nil {
		symbol = ""
	}
	symbolPassword := strings.Join(passwordWords, symbol)
	return plainPassword, symbolPassword
}

func (pg *PasswordGenerator) mapSymbols(input []rune) []rune {
	for i, char := range input {
		if replacement, ok := pg.values.SYMBOL_MAPPING[string(char)]; ok {
			input[i] = []rune(replacement)[0]
		}
	}
	return input
}

func (pg *PasswordGenerator) addRandomSymbols(pwd []rune, modifiedIndexes []int) ([]rune, []int) {
	count := rand.Intn(2) + 1
	for i := 0; i < count; {
		index := rand.Intn(len(pwd)-2) + 1

		if !contains(modifiedIndexes, index) {
			symbol, err := getSymbol(pg.values.INSIDE_SYMBOLS)
			if err != nil {
				log.Fatal().Err(err).Msg("Error getting INSIDE_SYMBOLS")
			}
			pwd[index] = []rune(string(symbol))[0]
			modifiedIndexes = append(modifiedIndexes, index)
			i++
		}
	}
	return pwd, modifiedIndexes
}

func (pg *PasswordGenerator) addRandomUppercase(pwd []rune, modifiedIndexes []int) ([]rune, []int) {
	count := rand.Intn(2) + 1
	for i := 0; i < count; {
		index := rand.Intn(len(pwd))

		if !contains(modifiedIndexes, index) {
			pwd[index] = []rune(strings.ToUpper(string(pwd[index])))[0]
			modifiedIndexes = append(modifiedIndexes, index)
			i++
		}
	}
	return pwd, modifiedIndexes
}

func (pg *PasswordGenerator) addRandomNumber(pwd []rune, modifiedIndexes []int) ([]rune, []int) {
	count := rand.Intn(2) + 1
	for i := 0; i < count; {
		index := rand.Intn(len(pwd))

		if !contains(modifiedIndexes, index) {
			pwd[index] = '0' + rune(rand.Intn(10))
			modifiedIndexes = append(modifiedIndexes, index)
			i++
		}
	}
	return pwd, modifiedIndexes
}

func (pg *PasswordGenerator) applyModifications(password []rune, modifiedIndexes []int) []rune {
	order := rand.Perm(3)
	for _, idx := range order {
		switch idx {
		case 0:
			password, modifiedIndexes = pg.addRandomUppercase(password, modifiedIndexes)
		case 1:
			password, modifiedIndexes = pg.addRandomSymbols(password, modifiedIndexes)
		case 2:
			password, modifiedIndexes = pg.addRandomNumber(password, modifiedIndexes)
		}
	}
	return password
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (pg *PasswordGenerator) GeneratePasswords(numPasswords int, minPasswordLength int, maxPasswordLength int, timeout ...time.Duration) ([]Password, error) {
	var wg sync.WaitGroup
	passwords := make([]Password, numPasswords)
	resultChan := make(chan Password, numPasswords)
	doneChan := make(chan bool)

	numGoroutines := (numPasswords + pg.values.PASSWORD_PER_ROUTINE - 1) / pg.values.PASSWORD_PER_ROUTINE

	// Set default timeout to 5 seconds if no timeout is provided
	timeoutDuration := 5 * time.Second
	if len(timeout) > 0 {
		timeoutDuration = timeout[0]
	}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		modifiedIndexes := make([]int, 0)

		numToGenerate := pg.values.PASSWORD_PER_ROUTINE
		if g == numGoroutines-1 {
			numToGenerate = numPasswords - (g * pg.values.PASSWORD_PER_ROUTINE)
		}
		go func(numToGenerate int, modifiedIndexes []int) {
			defer wg.Done()

			for i := 0; i < numToGenerate; i++ {
				original, modified := pg.generator(minPasswordLength, maxPasswordLength)
				modifiedRune := []rune(modified)
				xkcd := pg.applyModifications(modifiedRune, modifiedIndexes)
				xkcd = pg.mapSymbols(xkcd)
				length := len(xkcd)
				resultChan <- Password{Xkcd: string(xkcd), Original: original, Length: length}
			}
		}(numToGenerate, modifiedIndexes)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(doneChan)
	}()

	select {
	case <-doneChan:
		break
	case <-time.After(timeoutDuration):
		return nil, errors.New("Timeout")
	}
	i := 0
	for result := range resultChan {
		passwords[i] = result
		i++
	}

	return passwords, nil
}
