// Package auth отвечает за пароли и JWT-токены.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword возвращает bcrypt-хэш (соль уже внутри). bcrypt принимает
// не более 72 байт — длину ограничивает валидация на входе.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

// CheckPassword сравнивает пароль с хэшем.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
