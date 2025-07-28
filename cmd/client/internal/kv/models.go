package kv

type LoginPass struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Note struct {
	Text string `json:"text"`
}

type Card struct {
	Number string `json:"number"`
	Date   string `json:"date"`
	CVV    string `json:"cvv"`
}

type Binary struct {
	Data []byte `json:"data"`
}
