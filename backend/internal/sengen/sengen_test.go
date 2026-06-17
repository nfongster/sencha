package sengen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResponse_BasicPairs(t *testing.T) {
	text := `"저는 학생입니다"
"I am a student"
"물을 마시고 싶어요"
"I want to drink water"`

	pairs, err := parseResponse(text)
	assert.NoError(t, err)
	assert.Len(t, pairs, 2)
	assert.Equal(t, "저는 학생입니다", pairs[0].Korean)
	assert.Equal(t, "I am a student", pairs[0].English)
	assert.Equal(t, "물을 마시고 싶어요", pairs[1].Korean)
	assert.Equal(t, "I want to drink water", pairs[1].English)
}

func TestParseResponse_NoQuotes(t *testing.T) {
	text := `저는 학생입니다
I am a student
물을 마시고 싶어요
I want to drink water`

	pairs, err := parseResponse(text)
	assert.NoError(t, err)
	assert.Len(t, pairs, 2)
}

func TestParseResponse_EmptyLines(t *testing.T) {
	text := `"저는 학생입니다"
"I am a student"

"물을 마시고 싶어요"
"I want to drink water"`

	pairs, err := parseResponse(text)
	assert.NoError(t, err)
	assert.Len(t, pairs, 2)
}

func TestParseResponse_OddNumberOfLines(t *testing.T) {
	text := `"저는 학생입니다"
"I am a student"
"물을 마시고 싶어요"`

	pairs, err := parseResponse(text)
	assert.NoError(t, err)
	assert.Len(t, pairs, 1) // last line has no pair, dropped
}

func TestParseResponse_ASCIIQuotes(t *testing.T) {
	text := "\"저는 학생입니다\"\n\"I am a student\"\n\"물을 마시고 싶어요\"\n\"I want to drink water\""

	pairs, err := parseResponse(text)
	assert.NoError(t, err)
	assert.Len(t, pairs, 2)
}

func TestBuildPrompt_IncludesCount(t *testing.T) {
	prompt, err := buildPrompt(5)
	assert.NoError(t, err)
	assert.Contains(t, prompt, "5")
}

func TestBuildPrompt_IncludesGrammar(t *testing.T) {
	prompt, err := buildPrompt(10)
	assert.NoError(t, err)
	assert.Contains(t, prompt, "Subject-Object-Verb")
}

func TestBuildPrompt_IncludesVocab(t *testing.T) {
	prompt, err := buildPrompt(10)
	assert.NoError(t, err)
	assert.Contains(t, prompt, "학생")
	assert.Contains(t, prompt, "student")
}
