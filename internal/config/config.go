package config

// internal only
var (
	TVSRole      string = "tvs"
	LBRole       string = "lb"
	RedisRole    string = "redis"
	ObjectdbRole string = "objectdb"

	PossibleRoles = []string{LBRole, TVSRole, RedisRole, ObjectdbRole}
)

func RegionTxtRecord() string {
	return "regions." + APIDomain
}

func RegionNodeTxtRecord(region, role string) string {
	return role + "." + region + ".regions." + APIDomain
}

// internally set
var Home string

// settable in config file
var IsAPI bool
var IsRunner bool

var APIDomain string
var APIRootDomain string
var ToasterDomain string
var ToasterRootDomain string
var ToasterDomainSplitted []string

var LocalPrivateIP string

var LogLevel string // debug, warn, error, all - default is info

var Region string

var NodeDiscovery bool

var ObjectStorage struct {
	Name string // awss3

	AWSS3 struct {
		Region  string
		PubKey  string
		PrivKey string
		Bucket  string
	}

	LocalFS struct {
		Path string
	}
}

var EmailProvider struct {
	Name string // awsses

	AWSSES struct {
		Region      string
		PubKey      string
		PrivKey     string
		SourceEmail string
		ReplyTo     []string
	}
}

var Redis struct {
	IP []string // only if nodediscovery is set to false

	Username string
	Password string
	DB       int
}

var OBJECTDB struct {
	Kind string // sql, mongodb

	SQL struct {
		Syntax string // mysql, postgresql

		Hosts []string // only if nodediscovery is set to false

		Username string
		Password string
		DBName   string
	}
}

var DNSProvider struct {
	Name string

	ENV map[string]string // see golang lego dns providers example
}

var Runner struct {
	BTRFSMountPoint   string
	OverlayMountPoint string

	// Set one or the other
	UseUnmountedDisks bool
	BTRFSFile         string
	BTRSFileSize      int64

	NonRootUID int
	NonRootGID int

	NonRootUIDStr string
	NonRootGIDStr string

	ToasterPort string

	NetworkInterface string
}

var CertificateContactEmail string
