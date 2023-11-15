package passgen

import (
	cryptorand "crypto/rand"
	"errors"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"sync"
)

type PasswordGenerator struct {
	wordlist []string
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
	pg       *PasswordGenerator
	once     sync.Once
	wordList []string
)

func newPasswordGenerator(filename string) (*PasswordGenerator, error) {
	var err error
	once.Do(func() {
		wordList, err = loadWordsFromFile(filename)
		if err != nil {
			log.Fatalf("Error loading wordlist: %v", err)
		}
	})
	generator := &PasswordGenerator{
		wordlist: wordList,
	}

	return generator, nil
}

func loadWordsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("File %s not found.\n", filename)
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func getSymbol(pg *PasswordGenerator, symbols string) (string, error) {
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

func (pg *PasswordGenerator) generator(values Values) (string, string) {
	var symbol string
	passwordWords := make([]string, 3)
	totalLength := 0
	for totalLength < values.MIN_PASSWORD_LENGTH || totalLength > values.MAX_PASSWORD_LENGTH-2 {
		totalLength = 0
		for i := 0; i < 3; i++ {
			randomIndex, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(int64((len(pg.wordlist)))))
			passwordWords[i] = pg.wordlist[randomIndex.Int64()]
			totalLength += len(passwordWords[i])
		}
	}
	plainPassword := strings.Join(passwordWords, " ")
	symbol, err := getSymbol(pg, values.BETWEEN_SYMBOLS)
	if err != nil {
		symbol = ""
	}
	symbolPassword := strings.Join(passwordWords, symbol)
	return plainPassword, symbolPassword
}

func mapSymbols(input []rune, values Values) []rune {
	for i, char := range input {
		if replacement, ok := values.SYMBOL_MAPPING[string(char)]; ok {
			input[i] = []rune(replacement)[0]
		}
	}
	return input
}

func addRandomSymbols(pwd []rune, modifiedIndexes []int, values Values) ([]rune, []int) {
	count := rand.Intn(2) + 1
	for i := 0; i < count; {
		index := rand.Intn(len(pwd)-2) + 1

		if !contains(modifiedIndexes, index) {
			symbol, err := getSymbol(pg, values.INSIDE_SYMBOLS)
			if err != nil {
				log.Fatalf("INSIDE_SYMBOLS is empty")
			}
			pwd[index] = []rune(string(symbol))[0]
			modifiedIndexes = append(modifiedIndexes, index)
			i++
		}
	}
	return pwd, modifiedIndexes
}

func addRandomUppercase(pwd []rune, modifiedIndexes []int) ([]rune, []int) {
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

func addRandomNumber(pwd []rune, modifiedIndexes []int) ([]rune, []int) {
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

func applyModifications(password []rune, modifiedIndexes []int, values Values) []rune {
	order := rand.Perm(3)
	for _, idx := range order {
		switch idx {
		case 0:
			password, modifiedIndexes = addRandomUppercase(password, modifiedIndexes)
		case 1:
			password, modifiedIndexes = addRandomSymbols(password, modifiedIndexes, values)
		case 2:
			password, modifiedIndexes = addRandomNumber(password, modifiedIndexes)
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

func GeneratePasswords(filename string, numPasswords int, values Values) ([]Password, error) {
	pg, err := newPasswordGenerator(filename)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	passwords := make([]Password, numPasswords)
	resultChan := make(chan Password, numPasswords)

	numGoroutines := (numPasswords + values.PASSWORD_PER_ROUTINE - 1) / values.PASSWORD_PER_ROUTINE

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		modifiedIndexes := make([]int, 0)

		numToGenerate := values.PASSWORD_PER_ROUTINE
		if g == numGoroutines-1 {
			numToGenerate = numPasswords - g*values.PASSWORD_PER_ROUTINE
		}
		go func(numToGenerate int, modifiedIndexes []int) {
			defer wg.Done()

			for i := 0; i < numToGenerate; i++ {
				original, modified := pg.generator(values)
				modifiedRune := []rune(modified)
				xkcd := applyModifications(modifiedRune, modifiedIndexes, values)
				xkcd = mapSymbols(xkcd, values)
				length := len(xkcd)
				resultChan <- Password{Xkcd: string(xkcd), Original: original, Length: length}
			}
		}(numToGenerate, modifiedIndexes)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	i := 0
	for result := range resultChan {
		passwords[i] = result
		i++
	}

	return passwords, nil
}
