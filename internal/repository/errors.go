package repository

import "errors"

// ErrDuplicateEntry semnalează încălcarea unei constrângeri de unicitate în baza de date.
var ErrDuplicateEntry = errors.New("înregistrare duplicată")

// InUseError semnalează o ștergere refuzată de baza de date pentru că rândul are istoric
// (vezi „Ștergere” în README). Mesajul spune ce istoric îl ține.
type InUseError struct {
	Reason string
}

func (e *InUseError) Error() string { return e.Reason }

// ErrForeignKeyViolation semnalează o referință către o înregistrare inexistentă.
var ErrForeignKeyViolation = errors.New("referință inexistentă")
