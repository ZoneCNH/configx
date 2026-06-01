module github.com/ZoneCNH/configx

go 1.23

require (
	github.com/ZoneCNH/foundationx v0.0.0
	github.com/pelletier/go-toml/v2 v2.3.1
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/ZoneCNH/foundationx => ./internal/foundationx
