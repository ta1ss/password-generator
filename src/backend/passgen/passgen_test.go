package passgen

import (
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

var testValues Values

func TestMain(m *testing.M) {
	testValues, _ = loadValues()
	os.Exit(m.Run())
}

func loadValues() (Values, error) {
	var values Values

	yamlFile, err := os.ReadFile("../values/values.yaml")
	if err != nil {
		return values, err
	}

	err = yaml.Unmarshal(yamlFile, &values)
	if err != nil {
		return values, err
	}

	return values, nil
}

func TestApplyModifications(t *testing.T) {
	pwd := []rune("password")
	originalPwd := string(pwd)
	modifiedIndexes := []int{1, 3}
	result := applyModifications(pwd, modifiedIndexes, testValues)

	if len(pwd) != len(result) {
		t.Errorf("Expected length: %d, Got: %d", len(originalPwd), len(result))
	}

	if reflect.DeepEqual(originalPwd, result) {
		t.Errorf("Random Number not added")
	}
}

func TestMapSymbols(t *testing.T) {
	var input, expected []rune
	for k, v := range testValues.SYMBOL_MAPPING {
		input = append(input, []rune(k)...)
		expected = append(expected, []rune(v)...)
	}

	result := mapSymbols(input, testValues)

	if string(result) != string(expected) {
		t.Errorf("Expected: %s, Got: %s", string(expected), string(result))
	}
}

func TestAddRandomSymbols(t *testing.T) {
	pwd := []rune("password")
	originalPwd := string(pwd)

	modifiedIndexes := []int{1, 3}
	modifiedPwd, modifiedIndexes := addRandomSymbols(pwd, modifiedIndexes, testValues)

	if len(modifiedIndexes) < 3 || len(modifiedIndexes) > 4 {
		t.Errorf("Expected 3/4 modified indexes, Got: %d", len(modifiedIndexes))
	}

	if reflect.DeepEqual(originalPwd, modifiedPwd) {
		t.Errorf("Symbols not added")
	}
}

func TestAddRandomUppercase(t *testing.T) {
	pwd := []rune("password")
	originalPwd := string(pwd)
	modifiedIndexes := []int{1, 3}
	modifiedPwd, modifiedIndexes := addRandomUppercase(pwd, modifiedIndexes)

	if len(modifiedIndexes) < 3 || len(modifiedIndexes) > 4 {
		t.Errorf("Expected 3/4 modified indexes, Got: %d", len(modifiedIndexes))
	}

	if reflect.DeepEqual(originalPwd, modifiedPwd) {
		t.Errorf("Uppercase not added")
	}
}

func TestAddRandomNumber(t *testing.T) {
	pwd := []rune("password")
	originalPwd := string(pwd)
	modifiedIndexes := []int{1, 3}
	modifiedPwd, modifiedIndexes := addRandomNumber(pwd, modifiedIndexes)

	if len(modifiedIndexes) < 3 || len(modifiedIndexes) > 4 {
		t.Errorf("Expected 3/4 modified indexes, Got: %d", len(modifiedIndexes))
	}

	if reflect.DeepEqual(originalPwd, modifiedPwd) {
		t.Errorf("Random Number not added")
	}
}
