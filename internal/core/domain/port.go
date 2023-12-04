package domain

type Port struct {
	Id          string
	Name        string
	City        string
	Country     string
	Alias       []string //since both of these are empty i am leaving
	Regions     []string //same as above
	Coordinates []float64
	Province    string
	Timezone    string
	UNLOCs      []string
	Code        string
}
