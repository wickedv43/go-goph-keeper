package kv

// LoginPass represents a pair of login and password.
type LoginPass struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Note represents a plain text note.
type Note struct {
	Text string `json:"text"`
}

// Card represents credit card information.
type Card struct {
	Number string `json:"number"`
	Date   string `json:"date"`
	CVV    string `json:"cvv"`
}

// Binary represents arbitrary binary data.
type Binary struct {
	Data []byte `json:"data"`
}
