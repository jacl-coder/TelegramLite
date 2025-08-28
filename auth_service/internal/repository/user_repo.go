package repository

import (
    "context"
    "errors"
    "time"
)

type User struct {
    ID           int64
    Username     string
    PasswordHash string
    Role         string
    CreatedAt    time.Time
}

type UserRepo interface {
    Create(ctx context.Context, u *User) error
    GetByUsername(ctx context.Context, username string) (*User, error)
    GetByID(ctx context.Context, userID int64) (*User, error)
}

type userRepo struct {
    pg *PG
}

func NewUserRepo(pg *PG) UserRepo {
    return &userRepo{pg: pg}
}

func (r *userRepo) Create(ctx context.Context, u *User) error {
    var id int64
    err := r.pg.Pool.QueryRow(ctx,
        `INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) RETURNING id`,
        u.Username, u.PasswordHash, u.Role).Scan(&id)
    if err != nil {
        return err
    }
    u.ID = id
    return nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
    u := &User{}
    row := r.pg.Pool.QueryRow(ctx,
        `SELECT id, username, password_hash, role, created_at FROM users WHERE username=$1`, username)
    if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
        return nil, errors.New("not found")
    }
    return u, nil
}

func (r *userRepo) GetByID(ctx context.Context, userID int64) (*User, error) {
    u := &User{}
    row := r.pg.Pool.QueryRow(ctx,
        `SELECT id, username, password_hash, role, created_at FROM users WHERE id=$1`, userID)
    if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
        return nil, errors.New("not found")
    }
    return u, nil
}

