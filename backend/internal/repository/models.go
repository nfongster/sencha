package repository

type Phase struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}

type Level struct {
	Number       int    `json:"number"`
	PhaseNumber  int    `json:"phase_number"`
	GrammarMD    string `json:"grammar_md"`
	ExceptionsMD string `json:"exceptions_md"`
}

type VocabEntry struct {
	Korean   string `json:"korean"`
	English  string `json:"english"`
	Category string `json:"category"`
}

type Sentence struct {
	LevelNumber int
	Korean      string
	English     string
}

type LevelData struct {
	GrammarMD    string
	Vocab        []VocabEntry
	ExceptionsMD string
}
