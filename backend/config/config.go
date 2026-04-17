package config

type Server struct {
	Port int    `mapstructure:"port" yaml:"port"`
	Mode string `mapstructure:"mode" yaml:"mode"`
}

type MySQL struct {
	Path         string `mapstructure:"path" yaml:"path"`
	Port         int    `mapstructure:"port" yaml:"port"`
	Config       string `mapstructure:"config" yaml:"config"`
	DBName       string `mapstructure:"db-name" yaml:"db-name"`
	Username     string `mapstructure:"username" yaml:"username"`
	Password     string `mapstructure:"password" yaml:"password"`
	MaxIdleConns int    `mapstructure:"max-idle-conns" yaml:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns" yaml:"max-open-conns"`
	LogMode      bool   `mapstructure:"log-mode" yaml:"log-mode"`
	LogZap       bool   `mapstructure:"log-zap" yaml:"log-zap"`
}

type JWT struct {
	SigningKey  string `mapstructure:"signing-key" yaml:"signing-key"`
	ExpiresTime int    `mapstructure:"expires-time" yaml:"expires-time"`
	BufferTime  int    `mapstructure:"buffer-time" yaml:"buffer-time"`
	Issuer      string `mapstructure:"issuer" yaml:"issuer"`
}

type Log struct {
	Level         string `mapstructure:"level" yaml:"level"`
	Format        string `mapstructure:"format" yaml:"format"`
	Prefix        string `mapstructure:"prefix" yaml:"prefix"`
	Director      string `mapstructure:"director" yaml:"director"`
	ShowLine      bool   `mapstructure:"show-line" yaml:"show-line"`
	EncodeLevel   string `mapstructure:"encode-level" yaml:"encode-level"`
	StacktraceKey string `mapstructure:"stacktrace-key" yaml:"stacktrace-key"`
	LogInConsole  bool   `mapstructure:"log-in-console" yaml:"log-in-console"`
}

type Config struct {
	Server Server `mapstructure:"server" yaml:"server"`
	Mysql  MySQL  `mapstructure:"mysql" yaml:"mysql"`
	Jwt    JWT    `mapstructure:"jwt" yaml:"jwt"`
	Log    Log    `mapstructure:"log" yaml:"log"`
}
