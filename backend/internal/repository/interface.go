package repository

type Repository interface {
	// Phases
	ListPhases() ([]Phase, error)
	CreatePhase(p Phase) error
	UpdatePhase(number int, name string) error
	DeletePhase(number int) error
	MaxPhaseNumber() (int, error)

	// Levels
	LevelsInPhase(phaseNumber int) ([]Level, error)
	Level(number int) (*Level, error)
	CreateLevel(l Level) error
	UpdateLevel(number int, grammarMD, exceptionsMD string) error
	DeleteLevel(number int) error
	MaxLevelNumber() (int, error)
	LevelsUpTo(number int) ([]Level, error)

	// Vocabulary
	VocabularyUpTo(levelNumber int) ([]VocabEntry, error)
	VocabularyForLevel(levelNumber int) ([]VocabEntry, error)
	AddVocabulary(levelNumber int, entries []VocabEntry) error
	SetVocabulary(levelNumber int, entries []VocabEntry) error

	// Sentences
	SaveSentences(sentences []Sentence) error

	// Convenience
	LoadLevelData(levelNumber int) (*LevelData, error)
}
