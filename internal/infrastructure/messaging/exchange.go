package messaging

type ExchangeConfig struct {
	Name string
	Type string
}

var EventsExchange = ExchangeConfig{
	Name: "events",
	Type: "topic",
}
