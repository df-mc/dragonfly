module github.com/df-mc/dragonfly

go 1.24.1

require (
	github.com/brentp/intintmap v0.0.0-20190211203843-30dc0ade9af9
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/df-mc/goleveldb v1.1.9
	github.com/df-mc/worldupgrader v1.0.19
	github.com/go-gl/mathgl v1.2.0
	github.com/go-jose/go-jose/v4 v4.1.0
	github.com/google/uuid v1.6.0
	github.com/pelletier/go-toml v1.9.5
	github.com/sandertv/gophertunnel v1.46.0
	github.com/segmentio/fasthash v1.0.3
	golang.org/x/exp v0.0.0-20250606033433-dcc06ee1d476
	golang.org/x/mod v0.25.0
	golang.org/x/text v0.26.0
	golang.org/x/tools v0.34.0
)

require (
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/muhammadmuzzammil1998/jsonc v1.0.0 // indirect
	github.com/sandertv/go-raknet v1.14.3-0.20250525005230-991ee492a907 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/sandertv/go-raknet => github.com/aerisnetwork/aeris-raknet v0.1.0

replace github.com/sandertv/gophertunnel => github.com/aerisnetwork/aeris-gophertunnel v0.2.8
