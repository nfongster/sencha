package repository

import (
	"context"
	"fmt"
	"strings"

	"sencha/backend/internal/repository/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	ctx     context.Context
}

func NewPostgres(ctx context.Context, connString string) (*PostgresRepository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &PostgresRepository{
		queries: db.New(pool),
		pool:    pool,
		ctx:     ctx,
	}, nil
}

func (r *PostgresRepository) Close() {
	r.pool.Close()
}

func (r *PostgresRepository) ListPhases() ([]Phase, error) {
	rows, err := r.queries.ListPhases(r.ctx)
	if err != nil {
		return nil, err
	}
	phases := make([]Phase, len(rows))
	for i, row := range rows {
		phases[i] = Phase{Number: int(row.Number), Name: row.Name}
	}
	return phases, nil
}

func (r *PostgresRepository) CreatePhase(p Phase) error {
	return r.queries.CreatePhase(r.ctx, db.CreatePhaseParams{
		Number: int32(p.Number),
		Name:   p.Name,
	})
}

func (r *PostgresRepository) MaxPhaseNumber() (int, error) {
	v, err := r.queries.MaxPhaseNumber(r.ctx)
	if err != nil {
		return 0, err
	}
	n, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for max phase number: %T", v)
	}
	return int(n), nil
}

func (r *PostgresRepository) LevelsInPhase(phaseNumber int) ([]Level, error) {
	rows, err := r.queries.LevelsInPhase(r.ctx, int32(phaseNumber))
	if err != nil {
		return nil, err
	}
	levels := make([]Level, len(rows))
	for i, row := range rows {
		levels[i] = Level{
			Number:       int(row.Number),
			PhaseNumber:  int(row.PhaseNumber),
			GrammarMD:    row.GrammarMd,
			ExceptionsMD: row.ExceptionsMd,
		}
	}
	return levels, nil
}

func (r *PostgresRepository) Level(number int) (*Level, error) {
	row, err := r.queries.GetLevel(r.ctx, int32(number))
	if err != nil {
		return nil, err
	}
	return &Level{
		Number:       int(row.Number),
		PhaseNumber:  int(row.PhaseNumber),
		GrammarMD:    row.GrammarMd,
		ExceptionsMD: row.ExceptionsMd,
	}, nil
}

func (r *PostgresRepository) CreateLevel(l Level) error {
	var exceptions pgtype.Text
	if l.ExceptionsMD != "" {
		exceptions = pgtype.Text{String: l.ExceptionsMD, Valid: true}
	}
	return r.queries.CreateLevel(r.ctx, db.CreateLevelParams{
		Number:       int32(l.Number),
		PhaseNumber:  int32(l.PhaseNumber),
		GrammarMd:    l.GrammarMD,
		ExceptionsMd: exceptions,
	})
}

func (r *PostgresRepository) MaxLevelNumber() (int, error) {
	v, err := r.queries.MaxLevelNumber(r.ctx)
	if err != nil {
		return 0, err
	}
	n, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for max level number: %T", v)
	}
	return int(n), nil
}

func (r *PostgresRepository) LevelsUpTo(number int) ([]Level, error) {
	rows, err := r.queries.LevelsUpTo(r.ctx, int32(number))
	if err != nil {
		return nil, err
	}
	levels := make([]Level, len(rows))
	for i, row := range rows {
		levels[i] = Level{
			Number:       int(row.Number),
			PhaseNumber:  int(row.PhaseNumber),
			GrammarMD:    row.GrammarMd,
			ExceptionsMD: row.ExceptionsMd,
		}
	}
	return levels, nil
}

func (r *PostgresRepository) VocabularyUpTo(levelNumber int) ([]VocabEntry, error) {
	rows, err := r.queries.VocabularyUpTo(r.ctx, int32(levelNumber))
	if err != nil {
		return nil, err
	}
	entries := make([]VocabEntry, len(rows))
	for i, row := range rows {
		entries[i] = VocabEntry{Korean: row.Korean, English: row.English}
	}
	return entries, nil
}

func (r *PostgresRepository) AddVocabulary(levelNumber int, entries []VocabEntry) error {
	params := make([]db.AddVocabularyParams, len(entries))
	for i, e := range entries {
		params[i] = db.AddVocabularyParams{
			LevelNumber: int32(levelNumber),
			Korean:      e.Korean,
			English:     e.English,
		}
	}
	_, err := r.queries.AddVocabulary(r.ctx, params)
	return err
}

func (r *PostgresRepository) SaveSentences(sentences []Sentence) error {
	params := make([]db.SaveSentencesParams, len(sentences))
	for i, s := range sentences {
		params[i] = db.SaveSentencesParams{
			LevelNumber: int32(s.LevelNumber),
			Korean:      s.Korean,
			English:     s.English,
		}
	}
	_, err := r.queries.SaveSentences(r.ctx, params)
	return err
}

func (r *PostgresRepository) LoadLevelData(levelNumber int) (*LevelData, error) {
	levels, err := r.LevelsUpTo(levelNumber)
	if err != nil {
		return nil, fmt.Errorf("loading levels up to %d: %w", levelNumber, err)
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("no levels found up to %d", levelNumber)
	}

	var grammarParts []string
	var exceptionsMD string
	for _, l := range levels {
		if l.GrammarMD != "" {
			grammarParts = append(grammarParts, l.GrammarMD)
		}
		if l.ExceptionsMD != "" && l.Number == levelNumber {
			exceptionsMD = l.ExceptionsMD
		}
	}

	vocab, err := r.VocabularyUpTo(levelNumber)
	if err != nil {
		return nil, fmt.Errorf("loading vocabulary up to %d: %w", levelNumber, err)
	}

	return &LevelData{
		GrammarMD:    strings.Join(grammarParts, "\n\n"),
		Vocab:        vocab,
		ExceptionsMD: exceptionsMD,
	}, nil
}
