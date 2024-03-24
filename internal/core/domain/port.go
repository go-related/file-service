package domain

type Port struct {
	Id          string
	Name        string
	City        string
	Country     string
	Alias       []string
	Regions     []string
	Coordinates []float64
	Province    string
	Timezone    string
	UNLOCs      []string
	Code        string
}
