package args

// ServerArgs is a struct that holds the arguments for the CLI
type ServerArgs struct {
	Port          int    `mapstructure:"PORT"`
	Kubeconfig    string `mapstructure:"KUBECONFIG"`
	SelectorLabel string `mapstructure:"SELECTOR_LABEL"`
}
