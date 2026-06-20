package repository

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type memoryRepo struct {
	mu         sync.RWMutex
	phases     map[int]Phase
	levels     map[int]Level
	vocab      []VocabEntry // level_number → entries
	sentences  []Sentence
	nextLevel  int
}

func NewMemory() Repository {
	return &memoryRepo{
		phases:    make(map[int]Phase),
		levels:    make(map[int]Level),
		vocab:     nil,
		sentences: nil,
		nextLevel: 1,
	}
}

func (r *memoryRepo) ListPhases() ([]Phase, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Phase
	for _, p := range r.phases {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func (r *memoryRepo) CreatePhase(p Phase) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.phases[p.Number]; exists {
		return fmt.Errorf("phase %d already exists", p.Number)
	}
	if p.Number < 1 {
		return fmt.Errorf("phase number must be >= 1")
	}
	maxPhase := 0
	for n := range r.phases {
		if n > maxPhase {
			maxPhase = n
		}
	}
	if p.Number > maxPhase+1 {
		return fmt.Errorf("cannot create phase %d, max existing phase is %d (can create at most %d)", p.Number, maxPhase, maxPhase+1)
	}
	r.phases[p.Number] = p
	return nil
}

func (r *memoryRepo) UpdatePhase(number int, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.phases[number]
	if !ok {
		return fmt.Errorf("phase %d not found", number)
	}
	p.Name = name
	r.phases[number] = p
	return nil
}

func (r *memoryRepo) DeletePhase(number int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.phases[number]; !ok {
		return fmt.Errorf("phase %d not found", number)
	}
	for ln, l := range r.levels {
		if l.PhaseNumber == number {
			delete(r.levels, ln)
		}
	}
	delete(r.phases, number)
	return nil
}

func (r *memoryRepo) DeleteLevel(number int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.levels[number]
	if !ok {
		return fmt.Errorf("level %d not found", number)
	}
	phaseNum := l.PhaseNumber
	delete(r.levels, number)
	for ln, level := range r.levels {
		if level.PhaseNumber == phaseNum && ln > number {
			level.Number = ln - 1
			delete(r.levels, ln)
			r.levels[ln-1] = level
		}
	}
	return nil
}

func (r *memoryRepo) MaxPhaseNumber() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	maxPhase := 0
	for n := range r.phases {
		if n > maxPhase {
			maxPhase = n
		}
	}
	return maxPhase, nil
}

func (r *memoryRepo) LevelsInPhase(phaseNumber int) ([]Level, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Level
	for _, l := range r.levels {
		if l.PhaseNumber == phaseNumber {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func (r *memoryRepo) Level(number int) (*Level, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l, ok := r.levels[number]
	if !ok {
		return nil, fmt.Errorf("level %d not found", number)
	}
	return &l, nil
}

func (r *memoryRepo) CreateLevel(l Level) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	maxLevel := 0
	for n := range r.levels {
		if n > maxLevel {
			maxLevel = n
		}
	}
	next := maxLevel + 1
	if l.Number != 0 && l.Number != next {
		return fmt.Errorf("level number must be %d (next sequential), got %d", next, l.Number)
	}
	if _, exists := r.levels[next]; exists {
		return fmt.Errorf("level %d already exists", next)
	}
	if _, exists := r.phases[l.PhaseNumber]; !exists {
		maxPhase := 0
		for n := range r.phases {
			if n > maxPhase {
				maxPhase = n
			}
		}
		if l.PhaseNumber > maxPhase+1 {
			return fmt.Errorf("cannot create level in phase %d, max existing phase is %d", l.PhaseNumber, maxPhase)
		}
		r.phases[l.PhaseNumber] = Phase{Number: l.PhaseNumber, Name: fmt.Sprintf("Phase %d", l.PhaseNumber)}
	}
	l.Number = next
	r.levels[next] = l
	return nil
}

func (r *memoryRepo) UpdateLevel(number int, grammarMD, exceptionsMD string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.levels[number]
	if !ok {
		return fmt.Errorf("level %d not found", number)
	}
	l.GrammarMD = grammarMD
	l.ExceptionsMD = exceptionsMD
	r.levels[number] = l
	return nil
}

func (r *memoryRepo) MaxLevelNumber() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	maxLevel := 0
	for n := range r.levels {
		if n > maxLevel {
			maxLevel = n
		}
	}
	return maxLevel, nil
}

func (r *memoryRepo) LevelsUpTo(number int) ([]Level, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Level
	for _, l := range r.levels {
		if l.Number <= number {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func (r *memoryRepo) VocabularyUpTo(levelNumber int) ([]VocabEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.vocab, nil
}

func (r *memoryRepo) AddVocabulary(levelNumber int, entries []VocabEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.vocab = append(r.vocab, entries...)
	return nil
}

func (r *memoryRepo) SaveSentences(sentences []Sentence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sentences = append(r.sentences, sentences...)
	return nil
}

func (r *memoryRepo) LoadLevelData(levelNumber int) (*LevelData, error) {
	levels, err := r.LevelsUpTo(levelNumber)
	if err != nil {
		return nil, err
	}
	var grammarParts []string
	var exceptions string
	for _, l := range levels {
		if l.GrammarMD != "" {
			grammarParts = append(grammarParts, l.GrammarMD)
		}
		if l.Number == levelNumber && l.ExceptionsMD != "" {
			exceptions = l.ExceptionsMD
		}
	}

	vocab, err := r.VocabularyUpTo(levelNumber)
	if err != nil {
		return nil, err
	}

	return &LevelData{
		GrammarMD:    strings.Join(grammarParts, "\n\n"),
		Vocab:        vocab,
		ExceptionsMD: exceptions,
	}, nil
}
