package model

type Analysis struct {
	Classification string        `json:"classification"`
	Explanations   []Explanation `json:"explanations"`
	Notes          []string      `json:"notes"`
}

type Explanation struct {
	ID          string   `json:"id"`
	Severity    string   `json:"severity"`
	Message     string   `json:"message"`
	Details     string   `json:"details"`
	Suggestions []string `json:"suggestions"`
}

type Record struct {
	Version  string    `json:"version"`
	Run      RunResult `json:"run"`
	Analysis Analysis  `json:"analysis"`
}
