package objectdb

import (
	"fmt"

	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectdb/sqldb"
	"github.com/toastate/toastcloud/internal/model"
)

var Client interface {
	CreateUser(usr *model.User) error
	UpdateUser(usr *model.User) error
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(userid string) (*model.User, error)
	UserExistsByEmail(email string) (bool, error)

	CreateToaster(toaster *model.Toaster) error
	UpdateToaster(toaster *model.Toaster) error
	GetUserToaster(userid, toasterid string) (*model.Toaster, error)
	ListUsertoasters(userid string) ([]*model.Toaster, error)
	CheckToasterOwnership(userid, toasterid string) (bool, error)
	DelToaster(userid, toasterid string) error

	CreateSubDomain(sub *model.SubDomain) error
	UpdateSubDomain(sub *model.SubDomain) error
	DeleteSubDomain(userid, subdomainid string) error
	GetLinkedSubDomains(toasterid string) ([]*model.SubDomain, error)
	ListUserSubdomains(userid string) ([]*model.SubDomain, error)
	CheckSubdomainOwnership(subname, userid string) (bool, error)
	GetSubDomain(userid, subdomainid string) (*model.SubDomain, error)
	DeleteToasterAllSubdomains(userid, toasterid string) error

	UpsertCertificate(cert *model.Certificate) error
	GetCertificate(domain string) (*model.Certificate, error)
	DelCertificate(domain string) error

	BlockEmail(email, data string) error
	IsEmailBlocked(email string) (bool, error)
}

func Init() error {
	var err error

	switch config.OBJECTDB.Kind {
	case "sql":
		Client, err = sqldb.NewClient()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("we only support sql as an objectdb backend for now")
	}

	return nil
}
