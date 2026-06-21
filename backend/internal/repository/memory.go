package repository

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
)

type memoryRepo struct {
	mu              sync.RWMutex
	phases          map[int]Phase
	levels          map[int]Level
	vocab           []VocabEntry // cumulative, for VocabularyUpTo
	vocabByLevel    map[int][]VocabEntry
	categories      map[string]bool
	sentencesByLevel map[int][]Sentence
	nextLevel       int
}

func NewMemory() Repository {
	return &memoryRepo{
		phases:           make(map[int]Phase),
		levels:           make(map[int]Level),
		vocab:            nil,
		vocabByLevel:     make(map[int][]VocabEntry),
		categories:       make(map[string]bool),
		sentencesByLevel: make(map[int][]Sentence),
		nextLevel:        1,
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

func (r *memoryRepo) UpdateLevel(number int, grammarMD string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.levels[number]
	if !ok {
		return fmt.Errorf("level %d not found", number)
	}
	l.GrammarMD = grammarMD
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

func (r *memoryRepo) VocabularyForLevel(levelNumber int) ([]VocabEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.vocabByLevel[levelNumber], nil
}

func (r *memoryRepo) SetVocabulary(levelNumber int, entries []VocabEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.vocabByLevel[levelNumber] = entries
	r.rebuildCategories()
	return nil
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
	r.vocabByLevel[levelNumber] = append(r.vocabByLevel[levelNumber], entries...)
	for _, e := range entries {
		cat := e.Category
		if cat == "" {
			cat = "uncategorized"
		}
		r.categories[cat] = true
	}
	return nil
}

func (r *memoryRepo) Categories() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []string
	for cat := range r.categories {
		out = append(out, cat)
	}
	sort.Strings(out)
	return out, nil
}

func (r *memoryRepo) rebuildCategories() {
	r.categories = make(map[string]bool)
	for _, entries := range r.vocabByLevel {
		for _, e := range entries {
			cat := e.Category
			if cat == "" {
				cat = "uncategorized"
			}
			r.categories[cat] = true
		}
	}
}

func (r *memoryRepo) SaveSentences(sentences []Sentence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range sentences {
		r.sentencesByLevel[s.LevelNumber] = append(r.sentencesByLevel[s.LevelNumber], s)
	}
	return nil
}

func (r *memoryRepo) SentencesForLevel(levelNumber int) ([]Sentence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Sentence, len(r.sentencesByLevel[levelNumber]))
	copy(out, r.sentencesByLevel[levelNumber])
	return out, nil
}

func (r *memoryRepo) CountSentencesForLevel(levelNumber int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sentencesByLevel[levelNumber]), nil
}

func (r *memoryRepo) DeleteSentencesForLevel(levelNumber int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sentencesByLevel, levelNumber)
	return nil
}

func (r *memoryRepo) RandomSentencesForLevel(levelNumber int, count int) ([]Sentence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pool := r.sentencesByLevel[levelNumber]
	if len(pool) == 0 {
		return nil, nil
	}
	indices := rand.Perm(len(pool))
	n := count
	if n > len(pool) {
		n = len(pool)
	}
	out := make([]Sentence, n)
	for i := 0; i < n; i++ {
		out[i] = pool[indices[i]]
	}
	return out, nil
}

func (r *memoryRepo) LoadLevelData(levelNumber int) (*LevelData, error) {
	l, err := r.Level(levelNumber)
	if err != nil {
		return nil, err
	}

	categories, err := r.Categories()
	if err != nil {
		return nil, err
	}

	r.mu.RLock()
	byCategory := make(map[string][]VocabEntry)
	for _, entries := range r.vocabByLevel {
		for _, v := range entries {
			cat := v.Category
			if cat == "" {
				cat = "uncategorized"
			}
			byCategory[cat] = append(byCategory[cat], v)
		}
	}
	r.mu.RUnlock()

	// Use all categories sorted, or "uncategorized" if none exist
	catNames := categories
	if len(catNames) == 0 {
		catNames = []string{"uncategorized"}
	}

	target := 50
	perCat := target / len(catNames)
	remainder := target % len(catNames)

	var selected []VocabEntry
	for i, cat := range catNames {
		pool := byCategory[cat]
		rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		n := perCat
		if i == 0 {
			n += remainder
		}
		if len(pool) < n {
			n = len(pool)
		}
		selected = append(selected, pool[:n]...)
	}

	rand.Shuffle(len(selected), func(i, j int) { selected[i], selected[j] = selected[j], selected[i] })

	return &LevelData{
		GrammarMD: l.GrammarMD,
		Vocab:     selected,
	}, nil
}
