module main

go 1.18

replace github.com/UshakovN/stock-predictor-service v0.0.0-20230414193523-7fa2be658f07 => ./../pkg

require (
	github.com/UshakovN/stock-predictor-service v0.0.0-20230414193523-7fa2be658f07
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/UshakovN/token-bucket v1.1.0 // indirect
	github.com/elastic/go-elasticsearch/v7 v7.17.7 // indirect
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
