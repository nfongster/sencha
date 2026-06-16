package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSession_DefaultDirection(t *testing.T) {
	s := NewSession()
	assert.NotEmpty(t, s.ID)
	assert.Equal(t, DirectionKoreanToEnglish, s.Direction)
	assert.Equal(t, 10, s.TotalCards)
	assert.Equal(t, 10, s.CardsRemaining)
	assert.False(t, s.SessionComplete)
	assert.Equal(t, 0, s.CurrentIndex)
	assert.Equal(t, 10, len(s.Cards))
	assert.Empty(t, s.GradeCounts)
}

func TestNewSession_RespectsDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
	}{
		{"korean-to-english", DirectionKoreanToEnglish},
		{"english-to-korean", DirectionEnglishToKorean},
		{"mixed", DirectionMixed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSession(WithDirection(tt.direction))
			assert.Equal(t, tt.direction, s.Direction)
		})
	}
}

func TestNewSession_CardsShuffled(t *testing.T) {
	s1 := NewSession()
	s2 := NewSession()
	same := true
	for i := range s1.Cards {
		if s1.Cards[i].ID != s2.Cards[i].ID {
			same = false
			break
		}
	}
	if same {
		t.Fatal("two consecutive shuffles produced identical order — extremely unlikely, re-run the test")
	}
}

func TestReveal_ReturnsFrontAndBack(t *testing.T) {
	s := NewSession(WithDirection(DirectionKoreanToEnglish))
	card, err := s.Reveal()
	assert.NoError(t, err)
	assert.NotZero(t, card.ID)
	assert.NotEmpty(t, card.Front)
	assert.NotEmpty(t, card.Back)
	assert.True(t, s.Revealed)
}

func containsHangul(s string) bool {
	for _, r := range s {
		if r >= 0xAC00 && r <= 0xD7AF {
			return true
		}
	}
	return false
}

func TestReveal_RespectsKoreanToEnglish(t *testing.T) {
	s := NewSession(WithDirection(DirectionKoreanToEnglish))
	card, err := s.Reveal()
	assert.NoError(t, err)
	assert.True(t, containsHangul(card.Front), "front should contain Korean Hangul")
	assert.False(t, containsHangul(card.Back), "back should not contain Korean Hangul")
}

func TestReveal_RespectsEnglishToKorean(t *testing.T) {
	s := NewSession(WithDirection(DirectionEnglishToKorean))
	card, err := s.Reveal()
	assert.NoError(t, err)
	assert.False(t, containsHangul(card.Front), "front should not contain Korean Hangul")
	assert.True(t, containsHangul(card.Back), "back should contain Korean Hangul")
}

func TestGrade_WithoutReveal_ReturnsError(t *testing.T) {
	s := NewSession()
	summary, err := s.Grade(GradePass)
	assert.Error(t, err)
	assert.Nil(t, summary)
}

func TestReveal_Grade_AdvancesCard(t *testing.T) {
	s := NewSession(WithDirection(DirectionKoreanToEnglish))
	card1, _ := s.Reveal()
	summary, err := s.Grade(GradePass)
	assert.NoError(t, err)
	assert.Equal(t, 9, summary.CardsRemaining)
	assert.False(t, summary.SessionComplete)

	card2, _ := s.Reveal()
	assert.NotEqual(t, card1.ID, card2.ID)
}

func TestGrade_TracksCounts(t *testing.T) {
	s := NewSession()
	s.Reveal()
	s.Grade(GradePass)
	s.Reveal()
	s.Grade(GradeHard)
	s.Reveal()
	s.Grade(GradeFail)
	assert.Equal(t, 1, s.GradeCounts[GradePass])
	assert.Equal(t, 1, s.GradeCounts[GradeHard])
	assert.Equal(t, 1, s.GradeCounts[GradeFail])
}

func TestDoubleReveal_ReturnsSameCard(t *testing.T) {
	s := NewSession()
	card1, _ := s.Reveal()
	card2, _ := s.Reveal()
	assert.Equal(t, card1.ID, card2.ID)
	assert.Equal(t, card1.Front, card2.Front)
	assert.Equal(t, card1.Back, card2.Back)
}

func TestSessionComplete_AfterAllCards(t *testing.T) {
	s := NewSession()
	s.Reveal()
	// grade 9 cards (indices 0-8), revealing the next each time
	for i := 0; i < 9; i++ {
		summary, err := s.Grade(GradePass)
		assert.NoError(t, err)
		assert.False(t, summary.SessionComplete)
		s.Reveal()
	}
	// grade the 10th card (index 9) — should complete
	summary, err := s.Grade(GradePass)
	assert.NoError(t, err)
	assert.True(t, summary.SessionComplete)
	assert.Equal(t, 0, summary.CardsRemaining)
	assert.Equal(t, 10, summary.GradeCounts[GradePass])
}

func TestReveal_AfterComplete_ReturnsError(t *testing.T) {
	s := NewSession()
	s.Reveal()
	for i := 0; i < 9; i++ {
		s.Grade(GradePass)
		s.Reveal()
	}
	s.Grade(GradePass) // 10th grade, session complete
	_, err := s.Reveal()
	assert.Error(t, err)
}

func TestGrade_AfterComplete_ReturnsError(t *testing.T) {
	s := NewSession()
	s.Reveal()
	for i := 0; i < 9; i++ {
		s.Grade(GradePass)
		s.Reveal()
	}
	s.Grade(GradePass) // 10th grade, session complete
	_, err := s.Grade(GradePass)
	assert.Error(t, err)
}

func TestInvalidGrade_ReturnsError(t *testing.T) {
	s := NewSession()
	s.Reveal()
	_, err := s.Grade(Grade("invalid"))
	assert.Error(t, err)
}

func TestCards_AllUnique(t *testing.T) {
	cards := HardCodedCards()
	ids := make(map[int]bool)
	for _, c := range cards {
		assert.False(t, ids[c.ID], "duplicate card ID: %d", c.ID)
		ids[c.ID] = true
	}
	assert.Equal(t, 10, len(cards))
}
