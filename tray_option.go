package remotray

type Option func(config *Config)

func WithTitle(title string) Option {
	return func(config *Config) {
		config.title = title
	}
}

func WithTooltip(tooltip string) Option {
	return func(config *Config) {
		config.tooltip = tooltip
	}
}
