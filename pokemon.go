package main

type Pokemon struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	BaseExp int    `json:"base_experience"`
	Height  int
	Weight  int
	Stats   []struct {
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
		BaseStat int `json:"base_stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}
