package handler

const ctxKeyDescUserId = "user_id"

type ctxKeyUserId struct{}

func (ctxKeyUserId) KeyDescription() string {
  return ctxKeyDescUserId
}
