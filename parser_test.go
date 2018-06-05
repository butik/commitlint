package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	cfg, err := readConfig()

	require.NoError(t, err)
	assert.Equal(t, 72, cfg.HeaderMaxLength)
	assert.Len(t, cfg.AllowedTypes, 2)
}

func TestCheckLength(t *testing.T) {
	longString := "long-string"
	msgType := "type1"
	parsed := Parsed{
		Header: &longString,
		Type:   &msgType,
	}

	config := LintConfig{
		HeaderMaxLength: 1,
		AllowedTypes:    map[string]struct{}{"type1": {}},
	}

	res := check(parsed, config)
	require.Len(t, res, 1)
	assert.Equal(t, "header should be less than 1, actual 11", res[0].Description)
}

func TestCheckAllowedTypes(t *testing.T) {
	longString := "long-string"
	goodType := "type1"
	parsed := Parsed{
		Header: &longString,
		Type:   &goodType,
	}

	config := LintConfig{
		HeaderMaxLength: 72,
		AllowedTypes:    map[string]struct{}{"type1": {}},
	}

	res := check(parsed, config)
	require.Len(t, res, 0)
}

func TestCheckShouldRejectBadTypes(t *testing.T) {
	longString := "long-string"
	badType := "bad"
	parsed := Parsed{
		Header: &longString,
		Type:   &badType,
	}

	config := LintConfig{
		HeaderMaxLength: 72,
		AllowedTypes:    map[string]struct{}{"type1": {}},
	}

	res := check(parsed, config)
	require.Len(t, res, 1)

	assert.Equal(t, "type should be on of: type1", res[0].Description)
}

func TestParserShouldParseHeader(t *testing.T) {
	text := "feat(nglist): Allow custom separator"
	assert.NotNil(t, parse(text).Header)
	assert.Equal(t, text, *parse(text).Header)
}

func TestParseShouldExtractHeaderParts(t *testing.T) {
	text := "feat(scope): broadcast $destroy event on scope destruction"

	parsed := parse(text)

	assert.NotNil(t, parse(text).Header)
	assert.Equal(t, text, *parsed.Header)

	assert.NotNil(t, parse(text).Type)
	assert.Equal(t, "feat", *parsed.Type)

	assert.NotNil(t, parse(text).Scope)
	assert.Equal(t, "scope", *parsed.Scope)

	assert.NotNil(t, parse(text).Subject)
	assert.Equal(t, "broadcast $destroy event on scope destruction", *parsed.Subject)
}

func TestParseShouldSetNullIfPartNotFound(t *testing.T) {
	text := "header"

	parsed := parse(text)

	assert.NotNil(t, parse(text).Header)
	assert.Equal(t, text, *parsed.Header)

	assert.Nil(t, parsed.Type)
	assert.Nil(t, parsed.Scope)
	assert.Nil(t, parsed.Subject)
}
