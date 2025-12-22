package repository

import (
	"context"
	"database/sql"
	"fmt"

	"busoptima/internal/model"

	"github.com/jmoiron/sqlx"
)

// UserRepository інтерфейс для роботи з користувачами
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateRole(ctx context.Context, userID, roleID int64) error
	GetUserPermissions(ctx context.Context, userID int64) ([]string, error)
}

// userRepository реалізація UserRepository
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository створює новий екземпляр репозиторію користувачів
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// Create додає нового користувача до бази даних
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID повертає користувача за його ідентифікатором
func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	var roleName, roleDescription sql.NullString

	query := `
		SELECT u.id, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, 
		       u.created_at, u.updated_at, r.name as role_name, r.description as role_description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt, &roleName, &roleDescription,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Set role if exists
	if roleName.Valid {
		user.Role = &model.Role{
			ID:          user.RoleID,
			Name:        roleName.String,
			Description: roleDescription.String,
		}
	}

	return &user, nil
}

// GetByEmail повертає користувача за його email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	var roleName, roleDescription sql.NullString

	query := `
		SELECT u.id, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, 
		       u.created_at, u.updated_at, r.name as role_name, r.description as role_description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1 AND u.is_active = true`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt, &roleName, &roleDescription,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Set role if exists
	if roleName.Valid {
		user.Role = &model.Role{
			ID:          user.RoleID,
			Name:        roleName.String,
			Description: roleDescription.String,
		}
	}

	return &user, nil
}

// GetAll повертає список всіх користувачів
func (r *userRepository) GetAll(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, 
		       u.created_at, u.updated_at, r.name as role_name, r.description as role_description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var roleName, roleDescription sql.NullString

		err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive,
			&user.CreatedAt, &user.UpdatedAt, &roleName, &roleDescription,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Set role if exists
		if roleName.Valid {
			user.Role = &model.Role{
				ID:          user.RoleID,
				Name:        roleName.String,
				Description: roleDescription.String,
			}
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// Update оновлює існуючого користувача
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET 
			email = $1, full_name = $2, role_id = $3, is_active = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		user.Email, user.FullName, user.RoleID, user.IsActive, user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %d not found", user.ID)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateRole оновлює роль користувача
func (r *userRepository) UpdateRole(ctx context.Context, userID, roleID int64) error {
	query := `
		UPDATE users SET role_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, roleID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

// GetUserPermissions повертає список дозволів користувача
func (r *userRepository) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	var permissions []string
	query := `
		SELECT p.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE u.id = $1 AND u.is_active = true`

	err := r.db.SelectContext(ctx, &permissions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}
