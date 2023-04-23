package domain

type SignUpInput struct {
  Email    string `json:"email"`
  FullName string `json:"full_name"`
  Password string `json:"password"`
}

type SignInInput struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}
