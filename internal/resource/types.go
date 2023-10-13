package resource

type InstanceType int

const (
	EcsInstanceType InstanceType = iota
	MySQLInstanceType
	RedisInstanceType
)

type InstanceInfo struct {
	Type          InstanceType `ini:"-" json:"-"`
	ID            string       `ini:"-" json:"-"`
	PublicAddress string       `ini:"host" json:"host"`
	PublicPort    uint16       `ini:"port" json:"port"`
	Username      string       `ini:"user" json:"user"`
	Password      string       `ini:"password" json:"password"`
}

type Cloud interface {
	GetRegion() string
	GetInstances() ([]InstanceInfo, error)
}
