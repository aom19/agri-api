package config

// Acest fișier expune funcțiile interne verificate de testele din config/test.
// Testele stau într-un pachet separat și văd doar ce e exportat. Aplicația nu folosește nimic de aici.

var (
	GetEnv    = getEnv
	GetEnvInt = getEnvInt
)
