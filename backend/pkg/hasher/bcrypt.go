package hasher

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher { return &BcryptHasher{} }

func (h *BcryptHasher) Hash(password string) (string, error) {
    b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(b), err
}

func (h *BcryptHasher) Verify(hash, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}