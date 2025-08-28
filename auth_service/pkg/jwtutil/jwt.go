package jwtutil

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
    secret []byte
    ttl    time.Duration
}

func NewHMAC(secret []byte, ttl time.Duration) *JWTManager {
    return &JWTManager{secret: secret, ttl: ttl}
}

func (j *JWTManager) Mint(userID int64, username, role string) (string, error) {
    claims := jwt.MapClaims{
        "sub":  userID,
        "name": username,
        "role": role,
        "exp":  time.Now().Add(j.ttl).Unix(),
        "iat":  time.Now().Unix(),
    }
    t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return t.SignedString(j.secret)
}

func (j *JWTManager) Parse(tokenStr string) (map[string]any, error) {
    tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return j.secret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid {
        m := make(map[string]any)
        for k, v := range claims {
            m[k] = v
        }
        return m, nil
    }
    return nil, errors.New("invalid token")
}

