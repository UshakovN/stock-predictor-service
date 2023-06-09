package domain

import "time"

type UserInfo struct {
  UserId    string    `json:"user_id"`
  Email     string    `json:"email"`
  FullName  string    `json:"full_name"`
  CreatedAt time.Time `json:"created_at"`
}

type Tokens struct {
  Access  string `json:"access_token"`
  Refresh string `json:"refresh_token"`
}
