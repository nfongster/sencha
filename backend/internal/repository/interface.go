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
	UpdateLevel(number int, grammarMD string) error
	DeleteLevel(number int) error
	MaxLevelNumber() (int, error)
	LevelsUpTo(number int) ([]Level, error)

	// Vocabulary
	VocabularyUpTo(levelNumber int) ([]VocabEntry, error)
	VocabularyForLevel(levelNumber int) ([]VocabEntry, error)
	AddVocabulary(levelNumber int, entries []VocabEntry) error
	SetVocabulary(levelNumber int, entries []VocabEntry) error

	// Categories
	Categories() ([]string, error)

	// Sentences
	SaveSentences(sentences []Sentence) error
	SentencesForLevel(levelNumber int) ([]Sentence, error)
	CountSentencesForLevel(levelNumber int) (int, error)
	DeleteSentencesForLevel(levelNumber int) error
	RandomSentencesForLevel(levelNumber int, count int) ([]Sentence, error)

	// Convenience
	LoadLevelData(levelNumber int) (*LevelData, error)
}
