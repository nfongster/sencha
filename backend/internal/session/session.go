package session

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type Direction string

const (
	DirectionKoreanToEnglish Direction = "korean-to-english"
	DirectionEnglishToKorean Direction = "english-to-korean"
	DirectionMixed           Direction = "mixed"
)

type Grade string

const (
	GradePass Grade = "pass"
	GradeHard Grade = "hard"
	GradeFail Grade = "fail"
)

var validGrades = map[Grade]bool{
	GradePass: true,
	GradeHard: true,
	GradeFail: true,
}

type SentencePair struct {
	Korean  string
	English string
}

type Card struct {
	ID    int    `json:"id"`
	Front string `json:"front"`
	Back  string `json:"back"`
}

type GradeSummary struct {
	CardsRemaining  int          `json:"cards_remaining"`
	SessionComplete bool         `json:"session_complete"`
	GradeCounts     map[Grade]int `json:"grade_counts"`
}

type Session struct {
	ID              string
	Cards           []Card
	CurrentIndex    int
	Direction       Direction
	TotalCards      int
	CardsRemaining  int
	SessionComplete bool
	Revealed        bool
	GradeCounts     map[Grade]int
	sentencePairs   []SentencePair
	mu              sync.Mutex
}

type SessionOption func(*Session)

func WithDirection(d Direction) SessionOption {
	return func(s *Session) {
		s.Direction = d
	}
}

func WithPairs(pairs []SentencePair) SessionOption {
	return func(s *Session) {
		s.sentencePairs = pairs
	}
}

var idCounter int

func generateID() string {
	idCounter++
	return fmt.Sprintf("session-%d", idCounter)
}

func NewSession(opts ...SessionOption) *Session {
	s := &Session{
		ID:          generateID(),
		Direction:   DirectionKoreanToEnglish,
		GradeCounts: make(map[Grade]int),
	}
	for _, opt := range opts {
		opt(s)
	}

	if s.sentencePairs != nil {
		shuffled := make([]SentencePair, len(s.sentencePairs))
		copy(shuffled, s.sentencePairs)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		s.Cards = pairsToCards(shuffled, s.Direction)
	} else {
		rawCards := HardCodedCards()
		shuffled := make([]rawCard, len(rawCards))
		copy(shuffled, rawCards)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		s.Cards = cardsFromRaw(shuffled, s.Direction)
	}

	s.CardsRemaining = len(s.Cards)
	s.TotalCards = len(s.Cards)
	return s
}

func pairsToCards(pairs []SentencePair, dir Direction) []Card {
	cards := make([]Card, len(pairs))
	for i, p := range pairs {
		switch dir {
		case DirectionKoreanToEnglish:
			cards[i] = Card{ID: i + 1, Front: p.Korean, Back: p.English}
		case DirectionEnglishToKorean:
			cards[i] = Card{ID: i + 1, Front: p.English, Back: p.Korean}
		case DirectionMixed:
			if rand.Intn(2) == 0 {
				cards[i] = Card{ID: i + 1, Front: p.Korean, Back: p.English}
			} else {
				cards[i] = Card{ID: i + 1, Front: p.English, Back: p.Korean}
			}
		}
	}
	return cards
}

func cardsFromRaw(raw []rawCard, dir Direction) []Card {
	cards := make([]Card, len(raw))
	for i, rc := range raw {
		switch dir {
		case DirectionKoreanToEnglish:
			cards[i] = Card{ID: rc.ID, Front: rc.Korean, Back: rc.English}
		case DirectionEnglishToKorean:
			cards[i] = Card{ID: rc.ID, Front: rc.English, Back: rc.Korean}
		case DirectionMixed:
			if rand.Intn(2) == 0 {
				cards[i] = Card{ID: rc.ID, Front: rc.Korean, Back: rc.English}
			} else {
				cards[i] = Card{ID: rc.ID, Front: rc.English, Back: rc.Korean}
			}
		}
	}
	return cards
}

var ErrSessionComplete = errors.New("session already complete")
var ErrCardNotRevealed = errors.New("card not yet revealed")
var ErrInvalidGrade = errors.New("invalid grade")

func (s *Session) currentCard() Card {
	return s.Cards[s.CurrentIndex]
}

func (s *Session) Reveal() (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.SessionComplete {
		return Card{}, ErrSessionComplete
	}
	s.Revealed = true
	return s.currentCard(), nil
}

func (s *Session) Grade(g Grade) (*GradeSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.SessionComplete {
		return nil, ErrSessionComplete
	}
	if !s.Revealed {
		return nil, ErrCardNotRevealed
	}
	if !validGrades[g] {
		return nil, ErrInvalidGrade
	}

	s.GradeCounts[g]++
	s.CurrentIndex++
	s.Revealed = false
	s.CardsRemaining = len(s.Cards) - s.CurrentIndex

	if s.CurrentIndex >= len(s.Cards) {
		s.SessionComplete = true
		s.CardsRemaining = 0
	}

	counts := make(map[Grade]int)
	for k, v := range s.GradeCounts {
		counts[k] = v
	}

	return &GradeSummary{
		CardsRemaining:  s.CardsRemaining,
		SessionComplete: s.SessionComplete,
		GradeCounts:     counts,
	}, nil
}
