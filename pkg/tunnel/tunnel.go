package tunnel

// Provider defines the standard actions every VPN backend must support.
type Provider interface {
	Connect() error
	Disconnect() error
	Status() (string, error)
}
