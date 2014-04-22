package gopsutil

type Load struct{}

type LoadAvg struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

func NewLoad() Load {
	l := Load{}
	return l
}
