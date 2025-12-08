package peripheral

// PeriphConfiger is an interface that defines the Config method
// that all peripherals must implement
type PeriphConfiger interface {
	Config()
}
