package domain

// Paginarea listelor se face în baza de date. O listă cerută fără `page` rămâne nepaginată (toate
// rândurile, ca tablou JSON), pentru dropdown-uri, hărți și procesele de fundal; cu `page`, API-ul
// întoarce o pagină: Page.

const (
	// DefaultPageSize este mărimea paginii când clientul cere `page` fără `page_size`.
	DefaultPageSize = 25
	// MaxPageSize limitează o pagină, ca o cerere să nu poată citi tot tabelul dintr-odată.
	MaxPageSize = 200
)

// PageRequest este pagina cerută; Page începe de la 1.
type PageRequest struct {
	Page     int
	PageSize int
}

// Offset este numărul de rânduri sărite înaintea paginii.
func (p PageRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Page este o pagină dintr-o listă, cu numărul total de rânduri care corespund filtrelor.
type Page[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NewPage construiește pagina; Items nu e niciodată null în JSON.
func NewPage[T any](items []T, total int, request PageRequest) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}
}

// ListFilter este căutarea text comună listelor (nume, cod etc., după entitate).
type ListFilter struct {
	Query string
}
