package config

type Config struct {
	DB Db
}

type Db struct {
	DbUserName string
	DbUserPassword string
	DbHost string
	DbPort string
	DbName string
}
func New() (*Config, error) {
	cfg := &Config{
		DB: Db{
			DbUserName: "latona",
			DbPort:"30000",
			DbHost: "localhost",
			DbName: "PeripheralDevice",
			DbUserPassword: "L@ton@2019!",
		},
	}
	//if err := envconfig.Process("", cfg); err != nil {
	//	return nil, err
	//}
	return cfg, nil
}
