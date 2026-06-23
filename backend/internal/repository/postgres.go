package repository

import (
	"context"
	"fmt"
	"strings"

	"sencha/backend/internal/repository/db"

	"github.com/jackc/pgx/v5"
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

func (r *PostgresRepository) DeletePhase(number int) error {
	tx, err := r.pool.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	if _, err := tx.Exec(r.ctx, `DELETE FROM sentences WHERE level_number IN (SELECT number FROM levels WHERE phase_number = $1)`, number); err != nil {
		return err
	}
	if _, err := tx.Exec(r.ctx, `DELETE FROM vocabulary WHERE level_number IN (SELECT number FROM levels WHERE phase_number = $1)`, number); err != nil {
		return err
	}
	if _, err := tx.Exec(r.ctx, `DELETE FROM levels WHERE phase_number = $1`, number); err != nil {
		return err
	}
	result, err := tx.Exec(r.ctx, `DELETE FROM phases WHERE number = $1`, number)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("phase %d not found", number)
	}
	return tx.Commit(r.ctx)
}

func (r *PostgresRepository) DeleteLevel(number int) error {
	tx, err := r.pool.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	var phaseNumber int32
	if err := tx.QueryRow(r.ctx, `SELECT phase_number FROM levels WHERE number = $1`, number).Scan(&phaseNumber); err != nil {
		return fmt.Errorf("level %d not found", number)
	}
	if _, err := tx.Exec(r.ctx, `DELETE FROM sentences WHERE level_number = $1`, number); err != nil {
		return err
	}
	if _, err := tx.Exec(r.ctx, `DELETE FROM vocabulary WHERE level_number = $1`, number); err != nil {
		return err
	}
	if _, err := tx.Exec(r.ctx, `DELETE FROM levels WHERE number = $1`, number); err != nil {
		return err
	}
	if _, err := tx.Exec(r.ctx, `UPDATE levels SET number = number - 1 WHERE phase_number = $1 AND number > $2`, phaseNumber, number); err != nil {
		return err
	}
	return tx.Commit(r.ctx)
}

func (r *PostgresRepository) UpdatePhase(number int, name string) error {
	return r.queries.UpdatePhase(r.ctx, db.UpdatePhaseParams{Number: int32(number), Name: name})
}

func (r *PostgresRepository) MaxPhaseNumber() (int, error) {
	v, err := r.queries.MaxPhaseNumber(r.ctx)
	if err != nil {
		return 0, err
	}
	n, ok := v.(int32)
	if !ok {
		n64, ok := v.(int64)
		if !ok {
			return 0, fmt.Errorf("unexpected type for max phase number: %T", v)
		}
		return int(n64), nil
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
			Number:      int(row.Number),
			PhaseNumber: int(row.PhaseNumber),
			GrammarMD:   row.GrammarMd,
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
		Number:      int(row.Number),
		PhaseNumber: int(row.PhaseNumber),
		GrammarMD:   row.GrammarMd,
	}, nil
}

func (r *PostgresRepository) CreateLevel(l Level) error {
	return r.queries.CreateLevel(r.ctx, db.CreateLevelParams{
		Number:      int32(l.Number),
		PhaseNumber: int32(l.PhaseNumber),
		GrammarMd:   l.GrammarMD,
	})
}

func (r *PostgresRepository) UpdateLevel(number int, grammarMD string) error {
	return r.queries.UpdateLevel(r.ctx, db.UpdateLevelParams{
		Number:    int32(number),
		GrammarMd: grammarMD,
	})
}

func (r *PostgresRepository) MaxLevelNumber() (int, error) {
	v, err := r.queries.MaxLevelNumber(r.ctx)
	if err != nil {
		return 0, err
	}
	n, ok := v.(int32)
	if !ok {
		n64, ok := v.(int64)
		if !ok {
			return 0, fmt.Errorf("unexpected type for max level number: %T", v)
		}
		return int(n64), nil
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
			Number:      int(row.Number),
			PhaseNumber: int(row.PhaseNumber),
			GrammarMD:   row.GrammarMd,
		}
	}
	return levels, nil
}

func (r *PostgresRepository) VocabularyForLevel(levelNumber int) ([]VocabEntry, error) {
	rows, err := r.queries.VocabularyForLevel(r.ctx, int32(levelNumber))
	if err != nil {
		return nil, err
	}
	entries := make([]VocabEntry, len(rows))
	for i, row := range rows {
		entries[i] = VocabEntry{Korean: row.Korean, English: row.English, Category: row.Category}
	}
	return entries, nil
}

func (r *PostgresRepository) SetVocabulary(levelNumber int, entries []VocabEntry) error {
	tx, err := r.pool.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	q := r.queries.WithTx(tx)
		if err := q.DeleteVocabularyForLevel(r.ctx, int32(levelNumber)); err != nil {
		return err
	}
	if len(entries) > 0 {
		params := make([]db.AddVocabularyParams, len(entries))
		for i, e := range entries {
			params[i] = db.AddVocabularyParams{
				LevelNumber: int32(levelNumber),
				Korean:      e.Korean,
				English:     e.English,
				Category:    e.Category,
			}
		}
		if _, err := q.AddVocabulary(r.ctx, params); err != nil {
			return err
		}
	}
	return tx.Commit(r.ctx)
}

func (r *PostgresRepository) VocabularyUpTo(levelNumber int) ([]VocabEntry, error) {
	rows, err := r.queries.VocabularyUpTo(r.ctx, int32(levelNumber))
	if err != nil {
		return nil, err
	}
	entries := make([]VocabEntry, len(rows))
	for i, row := range rows {
		entries[i] = VocabEntry{Korean: row.Korean, English: row.English, Category: row.Category}
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
			Category:    e.Category,
		}
	}
	_, err := r.queries.AddVocabulary(r.ctx, params)
	return err
}

func (r *PostgresRepository) Categories() ([]string, error) {
	rows, err := r.pool.Query(r.ctx, `SELECT DISTINCT category FROM vocabulary ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
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
	l, err := r.Level(levelNumber)
	if err != nil {
		return nil, fmt.Errorf("level %d not found: %w", levelNumber, err)
	}

	categories, err := r.Categories()
	if err != nil {
		return nil, fmt.Errorf("fetching categories: %w", err)
	}

	target := 50

	// If no categories exist (empty DB), fall back to flat random
	var rows pgx.Rows
	if len(categories) == 0 {
		rows, err = r.pool.Query(r.ctx, `SELECT korean, english, category FROM vocabulary ORDER BY RANDOM() LIMIT 50`)
	} else {
		perCat := target / len(categories)
		remainder := target % len(categories)

		var subQueries []string
		var args []int32
		argIdx := 1
		for i := range categories {
			n := perCat
			if i == 0 {
				n += remainder
			}
			subQueries = append(subQueries,
				fmt.Sprintf(`(SELECT korean, english, category FROM vocabulary WHERE category = $%d LIMIT $%d)`, argIdx, argIdx+1))
			args = append(args, int32(n))
			argIdx += 2
		}

		query := "SELECT * FROM (" + strings.Join(subQueries, " UNION ALL ") + ") sub"
		if len(subQueries) > 1 {
			query += " ORDER BY RANDOM()"
		}

		// Build args: [cat1, limit1, cat2, limit2, ...]
		catArgs := make([]interface{}, 0, len(args)*2)
		for i, cat := range categories {
			catArgs = append(catArgs, cat, args[i])
		}
		rows, err = r.pool.Query(r.ctx, query, catArgs...)
	}
	if err != nil {
		return nil, fmt.Errorf("sampling vocabulary by category: %w", err)
	}
	defer rows.Close()

	var vocab []VocabEntry
	for rows.Next() {
		var v VocabEntry
		if err := rows.Scan(&v.Korean, &v.English, &v.Category); err != nil {
			return nil, err
		}
		vocab = append(vocab, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &LevelData{
		GrammarMD: l.GrammarMD,
		Vocab:     vocab,
	}, nil
}
