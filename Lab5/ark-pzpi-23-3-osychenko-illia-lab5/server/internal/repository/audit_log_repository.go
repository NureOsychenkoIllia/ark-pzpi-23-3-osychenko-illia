package repository

import (
	"busoptima/internal/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// AuditLogRepository інтерфейс для роботи з журналом аудиту
type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	GetAll(ctx context.Context, filters map[string]any) ([]model.AuditLog, error)
	GetCount(ctx context.Context, filters map[string]any) (int64, error)
}

// auditLogRepository реалізація AuditLogRepository
type auditLogRepository struct {
	db *sqlx.DB
}

// NewAuditLogRepository створює новий екземпляр репозиторію аудиту
func NewAuditLogRepository(db *sqlx.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create створює новий запис в журналі аудиту
func (r *auditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, 
			old_values, new_values, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	// Маршалінг JSON значень
	oldValuesJSON, err := json.Marshal(log.OldValues)
	if err != nil {
		return fmt.Errorf("failed to marshal old values: %w", err)
	}

	newValuesJSON, err := json.Marshal(log.NewValues)
	if err != nil {
		return fmt.Errorf("failed to marshal new values: %w", err)
	}

	return r.db.QueryRowContext(ctx, query,
		log.UserID, log.Action, log.EntityType, log.EntityID,
		oldValuesJSON, newValuesJSON, log.IPAddress,
	).Scan(&log.ID, &log.CreatedAt)
}

// GetAll повертає записи аудиту з фільтрами та пагінацією
func (r *auditLogRepository) GetAll(ctx context.Context, filters map[string]any) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	query := `
		SELECT al.id, al.user_id, al.action, al.entity_type, al.entity_id,
			al.old_values, al.new_values, al.ip_address, al.created_at,
			u.email, u.full_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE 1=1`

	args := []any{}
	argIndex := 1

	// Додаємо фільтри
	if userID, ok := filters["user_id"]; ok {
		query += fmt.Sprintf(" AND al.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if action, ok := filters["action"]; ok {
		query += fmt.Sprintf(" AND al.action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	if entityType, ok := filters["entity_type"]; ok {
		query += fmt.Sprintf(" AND al.entity_type = $%d", argIndex)
		args = append(args, entityType)
		argIndex++
	}

	if dateFrom, ok := filters["date_from"]; ok {
		query += fmt.Sprintf(" AND al.created_at >= $%d", argIndex)
		args = append(args, dateFrom)
		argIndex++
	}

	if dateTo, ok := filters["date_to"]; ok {
		query += fmt.Sprintf(" AND al.created_at <= $%d", argIndex)
		args = append(args, dateTo)
		argIndex++
	}

	query += " ORDER BY al.created_at DESC"

	// Пагінація
	if limit, ok := filters["limit"]; ok {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	} else {
		query += " LIMIT 100" // За замовчуванням
	}

	if offset, ok := filters["offset"]; ok {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
		argIndex++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var log model.AuditLog

		// Використовуємо nullable типи для LEFT JOIN полів
		var userEmail, userFullName *string
		var oldValuesBytes, newValuesBytes []byte

		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.EntityType, &log.EntityID,
			&oldValuesBytes, &newValuesBytes, &log.IPAddress, &log.CreatedAt,
			&userEmail, &userFullName,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Розбираємо JSON значення
		if oldValuesBytes != nil {
			if err := json.Unmarshal(oldValuesBytes, &log.OldValues); err != nil {
				log.OldValues = make(map[string]any)
			}
		} else {
			log.OldValues = make(map[string]any)
		}

		if newValuesBytes != nil {
			if err := json.Unmarshal(newValuesBytes, &log.NewValues); err != nil {
				log.NewValues = make(map[string]any)
			}
		} else {
			log.NewValues = make(map[string]any)
		}

		// Встановлюємо користувача тільки якщо є дані користувача
		if log.UserID != nil && userEmail != nil {
			user := model.User{
				ID:       *log.UserID,
				Email:    *userEmail,
				FullName: *userFullName,
			}
			log.User = &user
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetCount повертає загальну кількість записів аудиту з фільтрами
func (r *auditLogRepository) GetCount(ctx context.Context, filters map[string]any) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM audit_logs al
		WHERE 1=1`

	args := []any{}
	argIndex := 1

	// Додаємо ті ж фільтри, що й в GetAll (без пагінації)
	if userID, ok := filters["user_id"]; ok {
		query += fmt.Sprintf(" AND al.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if action, ok := filters["action"]; ok {
		query += fmt.Sprintf(" AND al.action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	if entityType, ok := filters["entity_type"]; ok {
		query += fmt.Sprintf(" AND al.entity_type = $%d", argIndex)
		args = append(args, entityType)
		argIndex++
	}

	if dateFrom, ok := filters["date_from"]; ok {
		query += fmt.Sprintf(" AND al.created_at >= $%d", argIndex)
		args = append(args, dateFrom)
		argIndex++
	}

	if dateTo, ok := filters["date_to"]; ok {
		query += fmt.Sprintf(" AND al.created_at <= $%d", argIndex)
		args = append(args, dateTo)
		argIndex++
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get audit logs count: %w", err)
	}

	return count, nil
}
