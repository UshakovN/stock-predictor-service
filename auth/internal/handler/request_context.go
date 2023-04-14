package handler

const ctxKeyDescUserId = "user_id"

type UserIdKey struct{}

func (k UserIdKey) KeyDescription() string {
  return ctxKeyDescUserId
}
