package publisher

type Publisher interface {
	Headers(headers []string) error
	Row(data []string) error
	Open() error
	Close()
}
