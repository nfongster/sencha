package repository

type Phase struct {
	Number int
	Name   string
}

type Level struct {
	Number       int
	PhaseNumber  int
	GrammarMD    string
	ExceptionsMD string
}

type VocabEntry struct {
	Korean  string
	English string
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
