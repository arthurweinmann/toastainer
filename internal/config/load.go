package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	AWSS3  loadedConfigAWSS3  `json:"aws_s3" yaml:"aws_s3"`
	AWSSES loadedConfigAWSSES `json:"aws_ses" yaml:"aws_ses"`

	Redis loadedConfigRedis `json:"redis" yaml:"redis"`

	SQLDB loadedConfigSQLDB `json:"sqldb" yaml:"sqldb"`

	DNSProvider loadedConfigDNSPROVIDER `json:"dns_provider" yaml:"dns_provider"`

	Runner loadedConfigRunner `json:"runner" yaml:"runner"`
}

type loadedConfigAWSS3 struct {
	Region  string `json:"aws_region" yaml:"aws_region"`
	PubKey  string `json:"public_key" yaml:"public_key"`
	PrivKey string `json:"private_key" yaml:"private_key"`
	Bucket  string `json:"bucket" yaml:"bucket"`
}

type loadedConfigAWSSES struct {
	Region      string   `json:"aws_region" yaml:"aws_region"`
	PubKey      string   `json:"public_key" yaml:"public_key"`
	PrivKey     string   `json:"private_key" yaml:"private_key"`
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
}

func LoadConfig(path string) error {
	lc := &loadedConfig{}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	switch filepath.Ext(path) {
	case ".json":
		err = json.Unmarshal(b, lc)
		if err != nil {
			return err
		}
	case ".yml":
		err = yaml.Unmarshal(b, lc)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported config file extension %s, toastcloud supports .json and .yml", filepath.Ext(path))
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
		return fmt.Errorf("you must configure one object storage")
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
	DNSProvider.ENV = lc.DNSProvider.ENV

	Runner.BTRFSMountPoint = lc.Runner.BTRFSMountPoint
	Runner.OverlayMountPoint = lc.Runner.OverlayMountPoint
	Runner.UseUnmountedDisks = lc.Runner.UseUnmountedDisks
	Runner.BTRFSFile = lc.Runner.BTRFSFile
	Runner.BTRSFileSize = lc.Runner.BTRSFileSize
	Runner.NonRootUID = lc.Runner.NonRootUID
	Runner.NonRootGID = lc.Runner.NonRootGID

	return nil
}
