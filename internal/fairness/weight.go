package fairness

type Tier struct {
	name     string
	capacity float64
	baseRate float64
	weight   int64
}

var Tiers = map[string]Tier{
	"free":       {name: "free", capacity: 100, baseRate: 100, weight: 1},
	"pro":        {name: "pro", capacity: 5000, baseRate: 1000, weight: 5},
	"enterprise": {name: "enterprise", capacity: 50000, baseRate: 10000, weight: 10},
}
