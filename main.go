package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

type Parsed struct {
	Type    *string
	Scope   *string
	Header  *string
	Subject *string
}

type LintConfig struct {
	HeaderMaxLength int
	AllowedTypes    map[string]struct{}
}

type LintError struct {
	Description string
}

func parse(text string) Parsed {
	res := Parsed{
		Header: &text,
	}

	r, _ := regexp.Compile(`^(\w*)(?:\(([\w$.\-* ]*)\))?: (.*)$`)

	if r.MatchString(text) {
		submatch := r.FindStringSubmatch(text)

		if len(submatch) > 1 {
			res.Type = &submatch[1]
		}
		if len(submatch) > 2 {
			res.Scope = &submatch[2]
		}
		if len(submatch) > 3 {
			res.Subject = &submatch[3]
		}
	}
	return res
}

func check(parsed Parsed, config LintConfig) []LintError {
	errors := make([]LintError, 0)

	if parsed.Header == nil {
		lintError := LintError{
			Description: "header should be non-empty",
		}
		errors = append(errors, lintError)
	} else if len(*parsed.Header) > config.HeaderMaxLength {
		lintError := LintError{
			Description: fmt.Sprintf("header should be less than %v, actual %v", config.HeaderMaxLength, len(*parsed.Header)),
		}
		errors = append(errors, lintError)
	}

	if parsed.Type == nil {
		lintError := LintError{
			Description: "type should be non-empty",
		}
		errors = append(errors, lintError)
	} else {
		if _, ok := config.AllowedTypes[*parsed.Type]; !ok {
			allowedTypes := make([]string, 0, len(config.AllowedTypes))
			for allowedType := range config.AllowedTypes {
				allowedTypes = append(allowedTypes, allowedType)
			}
			lintError := LintError{
				Description: fmt.Sprintf("type should be on of: %v", strings.Join(allowedTypes, ", ")),
			}
			errors = append(errors, lintError)
		}
	}

	return errors
}

type internalConfig struct {
	HeaderMaxLength int      `json:"header-max-length"`
	AllowedTypes    []string `json:"types"`
}

func readConfig() (LintConfig, error) {
	lintConfig := LintConfig{}

	file, err := os.Open(".commitlint")
	if err != nil {
		return lintConfig, fmt.Errorf("config .commitlint not found %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := internalConfig{}
	err = decoder.Decode(&configuration)
	if err != nil {
		if err != nil {
			return lintConfig, fmt.Errorf("error decode config %v", err)
		}
	}

	allowedTypes := make(map[string]struct{})
	for _, v := range configuration.AllowedTypes {
		allowedTypes[v] = struct{}{}
	}

	lintConfig.HeaderMaxLength = configuration.HeaderMaxLength
	lintConfig.AllowedTypes = allowedTypes

	return lintConfig, nil
}

func main() {
	errorColor := color.New(color.FgRed)
	neutralColor := color.New(color.Bold)
	goodColor := color.New(color.FgGreen, color.Bold)

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		errorColor.Println(err)
	}

	cfg, err := readConfig()
	if err != nil {
		errorColor.Println(err)
	}
	parsed := parse(text[:len(text)-1])
	lints := check(parsed, cfg)

	fmt.Print("⧗\tinput: ")
	neutralColor.Print(text)

	if len(lints) == 0 {
		goodColor.Print("✔")
		neutralColor.Println("\tAll ok!")
	} else {
		fmt.Println()
		for _, lintError := range lints {
			errorColor.Print("✖\t")
			errorColor.Println(lintError.Description)
		}
		errorColor.Print("✖\t")
		neutralColor.Printf("Found %v problems\n", len(lints))
	}
}
