module main

go 1.19

replace github.com/UshakovN/stock-predictor-service v0.0.0-20230414193523-7fa2be658f07 => ./../common-pkg

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/UshakovN/stock-predictor-service v0.0.0-20230414193523-7fa2be658f07
	github.com/jackc/pgconn v1.14.0
	github.com/jackc/pgx/v4 v4.18.1
	github.com/rabbitmq/amqp091-go v1.8.0
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.2 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
