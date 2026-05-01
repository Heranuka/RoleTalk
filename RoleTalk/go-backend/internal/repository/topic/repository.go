// Package topic implements the database repository for managing roleplay scenarios and user interactions.
package topic

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
	"go-backend/pkg/transactor"
)

var tracer = otel.Tracer("internal/repository/topic")

// Repository handles PostgreSQL operations for Topic entities with integrated observability.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates and returns a new Topic repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create inserts a new topic into the database and returns its generated UUID.
func (r *Repository) Create(ctx context.Context, t *domain.Topic) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "Topic.Repository.Create")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO topics (
			author_id, title, description, emoji, 
			difficulty_level, is_official, likes_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id uuid.UUID
	err := db.QueryRow(ctx, q,
		t.AuthorID, t.Title, t.Description, t.Emoji,
		t.DifficultyLevel, t.IsOfficial, t.LikesCount,
	).Scan(&id)

	if err != nil {
		r.handleInternalError(ctx, err, "failed to insert topic")
		return uuid.Nil, r.handlePostgresError(err, "create topic")
	}

	return id, nil
}

// GetOfficial retrieves all scenarios curated by the system.
func (r *Repository) GetOfficial(ctx context.Context) ([]*domain.Topic, error) {
	ctx, span := tracer.Start(ctx, "Topic.Repository.GetOfficial")
	defer span.End()

	const q = `
		SELECT id, author_id, title, description, emoji, difficulty_level, is_official, likes_count 
		FROM topics 
		WHERE is_official = true
		ORDER BY created_at DESC
	`

	return r.queryTopics(ctx, q)
}

// GetByID retrieves a single roleplay scenario from the database by its unique identifier.
// It returns ErrTopicNotFound if the record does not exist.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error) {
	// 1. Start a tracing span for Tempo
	ctx, span := tracer.Start(ctx, "Topic.Repository.GetByID")
	defer span.End()

	// 2. Get the database instance (from context or pool)
	db := transactor.GetDB(ctx, r.db)

	// 3. Define the query with explicit columns (essential for production stability)
	const q = `
		SELECT 
			id, 
			author_id, 
			title, 
			description, 
			emoji, 
			difficulty_level, 
			is_official, 
			likes_count,
			created_at
		FROM topics
		WHERE id = $1
	`

	var t domain.Topic
	err := db.QueryRow(ctx, q, id).Scan(
		&t.ID,
		&t.AuthorID,
		&t.Title,
		&t.Description,
		&t.Emoji,
		&t.DifficultyLevel,
		&t.IsOfficial,
		&t.LikesCount,
		&t.CreatedAt, // Ensure CreatedAt is present in your domain.Topic struct
	)

	if err != nil {
		// 4. Handle "Not Found" case
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTopicNotFound
		}

		// 5. Log internal technical error for Loki
		r.handleInternalError(ctx, err, "failed to fetch topic by id")

		// 6. Return mapped error for the Service layer
		return nil, r.handlePostgresError(err, "get topic by id")
	}

	return &t, nil
}

// GetCommunity retrieves scenarios created by users with pagination and ranking.
func (r *Repository) GetCommunity(ctx context.Context, limit, offset int) ([]*domain.Topic, error) {
	ctx, span := tracer.Start(ctx, "Topic.Repository.GetCommunity")
	defer span.End()

	const q = `
		SELECT id, author_id, title, description, emoji, difficulty_level, is_official, likes_count 
		FROM topics 
		WHERE is_official = false
		ORDER BY likes_count DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	return r.queryTopics(ctx, q, limit, offset)
}

// AddLike inserts a record into topic_likes and increments the topic's like counter.
// It uses a transaction to ensure data consistency.
func (r *Repository) AddLike(ctx context.Context, userID, topicID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Topic.Repository.AddLike")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	// 1. Insert into topic_likes table
	const qInsertLike = `INSERT INTO topic_likes (user_id, topic_id) VALUES ($1, $2)`
	_, err := db.Exec(ctx, qInsertLike, userID, topicID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrLikeAlreadyExists
		}
		r.handleInternalError(ctx, err, "failed to insert topic_like")
		return r.handlePostgresError(err, "add like")
	}

	// 2. Increment likes_count in topics table
	const qUpdateCounter = `UPDATE topics SET likes_count = likes_count + 1 WHERE id = $1`
	res, err := db.Exec(ctx, qUpdateCounter, topicID)
	if err != nil {
		r.handleInternalError(ctx, err, "failed to increment topic counter")
		return r.handlePostgresError(err, "increment like counter")
	}

	if res.RowsAffected() == 0 {
		return ErrTopicNotFound
	}

	return nil
}

// RemoveLike deletes a record from topic_likes and decrements the topic's like counter.
func (r *Repository) RemoveLike(ctx context.Context, userID, topicID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Topic.Repository.RemoveLike")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	// 1. Delete from topic_likes table
	const qDeleteLike = `DELETE FROM topic_likes WHERE user_id = $1 AND topic_id = $2`
	res, err := db.Exec(ctx, qDeleteLike, userID, topicID)
	if err != nil {
		r.handleInternalError(ctx, err, "failed to delete topic_like")
		return r.handlePostgresError(err, "remove like")
	}

	if res.RowsAffected() == 0 {
		return ErrLikeNotFound
	}

	// 2. Decrement likes_count in topics table (ensure it doesn't go below 0)
	const qUpdateCounter = `UPDATE topics SET likes_count = GREATEST(0, likes_count - 1) WHERE id = $1`
	_, err = db.Exec(ctx, qUpdateCounter, topicID)
	if err != nil {
		r.handleInternalError(ctx, err, "failed to decrement topic counter")
		return r.handlePostgresError(err, "decrement like counter")
	}

	return nil
}

// Delete removes a topic record from the database.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Topic.Repository.Delete")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `DELETE FROM topics WHERE id = $1`

	res, err := db.Exec(ctx, q, id)
	if err != nil {
		r.handleInternalError(ctx, err, "failed to execute delete query")
		return r.handlePostgresError(err, "delete topic")
	}

	if res.RowsAffected() == 0 {
		return ErrTopicNotFound
	}

	return nil
}

// queryTopics is an internal helper to execute queries and scan multiple topic rows.
func (r *Repository) queryTopics(ctx context.Context, query string, args ...any) ([]*domain.Topic, error) {
	db := transactor.GetDB(ctx, r.db)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		r.handleInternalError(ctx, err, "failed to query topics")
		return nil, r.handlePostgresError(err, "query topics")
	}
	defer rows.Close()

	var topics []*domain.Topic
	for rows.Next() {
		t := &domain.Topic{}
		err := rows.Scan(
			&t.ID, &t.AuthorID, &t.Title, &t.Description,
			&t.Emoji, &t.DifficultyLevel, &t.IsOfficial, &t.LikesCount,
		)
		if err != nil {
			r.handleInternalError(ctx, err, "failed to scan topic row")
			return nil, r.handlePostgresError(err, "scan topics")
		}
		topics = append(topics, t)
	}

	return topics, nil
}

// handleInternalError records technical failures in the trace span and logs them via Zap.
func (r *Repository) handleInternalError(ctx context.Context, err error, message string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, message)

	log.Errorw(message, "error", err)
}

// handlePostgresError maps PostgreSQL-specific constraint violations to domain errors.
func (r *Repository) handlePostgresError(err error, operation string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			if pgErr.ConstraintName == "topics_author_id_fkey" {
				return ErrAuthorNotFound
			}
		case pgerrcode.UniqueViolation:
			if pgErr.ConstraintName == "topic_likes_pkey" || pgErr.ConstraintName == "topic_likes_user_id_topic_id_key" {
				return ErrLikeAlreadyExists
			}
			return ErrDuplicateTitle
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTopicNotFound
	}

	return fmt.Errorf("db %s failed: %w", operation, err)
}
