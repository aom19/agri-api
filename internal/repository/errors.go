package repository

import "errors"

// ErrDuplicateEntry semnalează încălcarea unei constrângeri de unicitate în baza de date.
var ErrDuplicateEntry = errors.New("înregistrare duplicată")
