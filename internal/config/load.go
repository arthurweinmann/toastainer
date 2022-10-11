package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type loadedConfig struct {
	IsAPI    bool `json:"is_api" yaml:"is_api"`
	IsRunner bool `json:"is_runner" yaml:"is_runner"`

	APIDomain         string `json:"api_domain" yaml:"api_domain"`
	APIRootDomain     string `json:"api_root_domain" yaml:"api_root_domain"`
	ToasterDomain     string `json:"toaster_domain" yaml:"toaster_domain"`
	ToasterRootDomain string `json:"toaster_root_domain" yaml:"toaster_root_domain"`

	CertificateContactEmail string `json:"ssl_certificate_contact_email" yaml:"ssl_certificate_contact_email"`

	LogLevel string `json:"log_level" yaml:"log_level"`

	NodeDiscovery  bool   `json:"node_discovery" yaml:"node_discovery"`
	LocalPrivateIP string `json:"node_local_private_network_ip" yaml:"node_local_private_network_ip"`
	Region         string `json:"region" yaml:"region"`

	AWSS3   loadedConfigAWSS3   `json:"aws_s3" yaml:"aws_s3"`
	LocalFS loadedConfigLocalFS `json:"local_filestorage" yaml:"local_filestorage"`

	AWSSES loadedConfigAWSSES `json:"aws_ses" yaml:"aws_ses"`

	Redis loadedConfigRedis `json:"redis" yaml:"redis"`

	SQLDB loadedConfigSQLDB `json:"sqldb" yaml:"sqldb"`

	DNSProvider loadedConfigDNSPROVIDER `json:"dns_provider" yaml:"dns_provider"`

	Runner loadedConfigRunner `json:"runner" yaml:"runner"`
}

type loadedConfigAWSS3 struct {
	Region  string `json:"region" yaml:"region"`
	PubKey  string `json:"access_key_id" yaml:"access_key_id"`
	PrivKey string `json:"secret_access_key" yaml:"secret_access_key"`
	Bucket  string `json:"bucket" yaml:"bucket"`
}

type loadedConfigLocalFS struct {
	Path string `json:"path" yaml:"path"`
}

type loadedConfigAWSSES struct {
	Region      string   `json:"region" yaml:"region"`
	PubKey      string   `json:"access_key_id" yaml:"access_key_id"`
	PrivKey     string   `json:"secret_access_key" yaml:"secret_access_key"`
	SourceEmail string   `json:"source_email" yaml:"source_email"`
	ReplyTo     []string `json:"reply_to" yaml:"reply_to"`
}

type loadedConfigRedis struct {
	IP       []string `json:"hosts" yaml:"hosts"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	DB       int      `json:"db_number" yaml:"db_number"`
}

type loadedConfigSQLDB struct {
	Syntax   string   `json:"syntax" yaml:"syntax"`
	Hosts    []string `json:"hosts" yaml:"hosts"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	DBName   string   `json:"db_name" yaml:"db_name"`
}

type loadedConfigDNSPROVIDER struct {
	Name string            `json:"name" yaml:"name"`
	ENV  map[string]string `json:"configuration" yaml:"configuration"` // see Lego https://go-acme.github.io/lego/dns/
}

type loadedConfigRunner struct {
	BTRFSMountPoint   string `json:"btrfs_mountpoint" yaml:"btrfs_mountpoint"`
	OverlayMountPoint string `json:"overlayfs_mountpoint" yaml:"overlayfs_mountpoint"`
	UseUnmountedDisks bool   `json:"use_unmounted_disks" yaml:"use_unmounted_disks"`
	BTRFSFile         string `json:"btrfs_filepath" yaml:"btrfs_filepath"`
	BTRSFileSize      int64  `json:"btrfs_filesize" yaml:"btrfs_filesize"`
	NonRootUID        int    `json:"non_root_uid" yaml:"non_root_uid"`
	NonRootGID        int    `json:"non_root_gid" yaml:"non_root_gid"`
	ToasterPort       int    `json:"toaster_port" yaml:"toaster_port"`
	NetworkInterface  string `json:"network_interface" yaml:"network_interface"`
}

func LoadConfig(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	if len(ext) > 1 {
		ext = ext[1:]
	}

	return LoadConfigBytes(b, ext)
}

// extension is either json or yml
func LoadConfigBytes(b []byte, extension string) error {
	lc := &loadedConfig{}
	var err error

	switch extension {
	case "json":
		err = json.Unmarshal(b, lc)
		if err != nil {
			return err
		}
	case "yml":
		err = yaml.Unmarshal(b, lc)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported config format %s, toastainer supports json and yml", extension)
	}

	IsAPI = lc.IsAPI
	IsRunner = lc.IsRunner

	APIRootDomain = lc.APIRootDomain
	APIDomain = lc.APIDomain
	ToasterDomain = lc.ToasterDomain
	ToasterRootDomain = lc.ToasterRootDomain
	ToasterDomainSplitted = strings.Split(ToasterDomain, ".")

	LocalPrivateIP = lc.LocalPrivateIP
	NodeDiscovery = lc.NodeDiscovery
	Region = lc.Region

	LogLevel = lc.LogLevel

	CertificateContactEmail = lc.CertificateContactEmail

	switch {
	case lc.AWSS3.PrivKey != "":
		ObjectStorage.Name = "awss3"
		ObjectStorage.AWSS3.Region = lc.AWSS3.Region
		ObjectStorage.AWSS3.PubKey = lc.AWSS3.PubKey
		ObjectStorage.AWSS3.PrivKey = lc.AWSS3.PrivKey
		ObjectStorage.AWSS3.Bucket = lc.AWSS3.Bucket

	case lc.LocalFS.Path != "":
		ObjectStorage.Name = "localfs"
		ObjectStorage.LocalFS.Path = lc.LocalFS.Path

	default:
		return fmt.Errorf("you must configure one object storage")
	}

	switch {
	case lc.AWSSES.PrivKey != "":
		EmailProvider.Name = "awsses"
		EmailProvider.AWSSES.Region = lc.AWSSES.Region
		EmailProvider.AWSSES.PubKey = lc.AWSSES.PubKey
		EmailProvider.AWSSES.PrivKey = lc.AWSSES.PrivKey
		EmailProvider.AWSSES.SourceEmail = lc.AWSSES.SourceEmail
		EmailProvider.AWSSES.ReplyTo = lc.AWSSES.ReplyTo

	default:
		return fmt.Errorf("you must configure one email provider")
	}

	Redis.IP = lc.Redis.IP
	Redis.Username = lc.Redis.Username
	Redis.Password = lc.Redis.Password
	Redis.DB = lc.Redis.DB

	switch {
	case lc.SQLDB.Syntax != "":
		OBJECTDB.Kind = "sql"
		OBJECTDB.SQL.Syntax = lc.SQLDB.Syntax
		OBJECTDB.SQL.Hosts = lc.SQLDB.Hosts
		OBJECTDB.SQL.Username = lc.SQLDB.Username
		OBJECTDB.SQL.Password = lc.SQLDB.Password
		OBJECTDB.SQL.DBName = lc.SQLDB.DBName

	default:
		return fmt.Errorf("you must configure one object database")
	}

	DNSProvider.Name = lc.DNSProvider.Name
	DNSProvider.ENV = map[string]string{}
	for k, v := range lc.DNSProvider.ENV {
		DNSProvider.ENV[strings.ToUpper(k)] = v
	}

	Runner.BTRFSMountPoint = lc.Runner.BTRFSMountPoint
	Runner.OverlayMountPoint = lc.Runner.OverlayMountPoint
	Runner.UseUnmountedDisks = lc.Runner.UseUnmountedDisks
	Runner.NetworkInterface = lc.Runner.NetworkInterface
	Runner.BTRFSFile = lc.Runner.BTRFSFile
	Runner.BTRSFileSize = lc.Runner.BTRSFileSize
	if lc.Runner.ToasterPort != 0 {
		Runner.ToasterPort = strconv.Itoa(lc.Runner.ToasterPort)
	}

	Runner.NonRootUID = lc.Runner.NonRootUID
	Runner.NonRootGID = lc.Runner.NonRootGID
	Runner.NonRootUIDStr = strconv.Itoa(lc.Runner.NonRootUID)
	Runner.NonRootGIDStr = strconv.Itoa(lc.Runner.NonRootGID)

	return checkConfig()
}

func checkConfig() error {
	if DNSProvider.Name == "" {
		return fmt.Errorf("you must provide a DNS provider for Toasters SSL certificates")
	}

	return nil
}
