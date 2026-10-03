package handlers

// Acest fișier expune funcțiile interne verificate de testele din handlers/test.
// Testele stau într-un pachet separat și văd doar ce e exportat. Aplicația nu folosește nimic de aici.

var (
	AuditID           = auditID
	StatusChange      = statusChange
	IsOperatorRequest = isOperatorRequest
	CurrentUserID     = currentUserID
	CurrentActorID    = currentActorID
	ValidationErrors  = validationErrors
)
