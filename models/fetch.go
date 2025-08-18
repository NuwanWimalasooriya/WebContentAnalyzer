package models


type Heading struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

type FetchResponse struct {
	Title           string   `json:"title"`
	HtmlVersion		string	 `json:"htmlVersion"`
	Headings        []Heading `json:"headings"`
	Links           []string `json:"links"`
	LoginDetected   bool     `json:"login_detected"`
	LoginIndicators []string `json:"login_indicators,omitempty"`
	Error           string   `json:"error,omitempty"`
}

type FetchRequest struct {
    URL string `json:"url"`
}
