package config

type Config struct {
	DBSettings      DBSettings      `yaml:"DBSettings"`
	ProgramSettings ProgramSettings `yaml:"ProgramSettings"`
}

type ProgramSettings struct {
	JiraURL           string `yaml:"jiraUrl"`
	ThreadCount       int    `yaml:"threadCount"`
	IssueInOneRequest int    `yaml:"issueInOneRequest"`
	MinTimeSleep      int    `yaml:"minTimeSleep"`
	MaxTimeSleep      int    `yaml:"maxTimeSleep"`
}

type DBSettings struct {
	User     string `yaml:"dbUser"`
	Password string `yaml:"dbPassword"`
	Host     string `yaml:"dbHost"`
	Port     int    `yaml:"dbPort"`
	Name     string `yaml:"dbName"`
}
