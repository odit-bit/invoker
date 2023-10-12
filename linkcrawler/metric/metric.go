package metric

type Counter interface {
	Add(float64)
}

type CounterFunc func(float64)
