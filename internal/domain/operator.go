package domain

import "strings"

type OperatorStatus string

const (
	OperatorStatusActive   OperatorStatus = "active"
	OperatorStatusInactive OperatorStatus = "inactive"
)

// PlaceholderEmailDomain este domeniul adreselor tehnice date operatorilor fără e-mail. Contul
// există (lucrările îi sunt asignate), dar nu se poate folosi până nu primește un e-mail real.
const PlaceholderEmailDomain = "fara-email.local"

// IsPlaceholderEmail spune dacă adresa este una tehnică, nu e-mailul real al persoanei.
func IsPlaceholderEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), "@"+PlaceholderEmailDomain)
}

// Operator este un utilizator cu rolul `operator`: contul (users) plus profilul (user_profiles).
// ID-ul este id-ul utilizatorului; statusul inactiv înseamnă cont dezactivat.
type Operator struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	// Email este gol când contul are doar o adresă tehnică.
	Email  string         `json:"email"`
	Notes  string         `json:"notes"`
	Status OperatorStatus `json:"status"`
}

// FullName este numele afișat: prenume + nume.
func (o Operator) FullName() string {
	return strings.TrimSpace(strings.Join(strings.Fields(o.FirstName+" "+o.LastName), " "))
}
