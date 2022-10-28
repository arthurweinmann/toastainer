package objectdb

import (
	"fmt"

	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectdb/sqldb"
	"github.com/toastate/toastainer/internal/model"
)

var Client interface {
	CreateUser(usr *model.User) error
	UpdateUser(usr *model.User) error
	GetUserByEmail(email string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByID(userid string) (*model.User, error)
	RangeUsers(limit int, cursor string) (string, bool, []model.User, error)
	UserExistsByEmail(email string) (bool, error)
	UserExistsByUsername(username string) (bool, error)
	DelUser(userid string) error

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
	UnlinkAllSubdomainsFromToaster(userid, toasterid string) error
	DeleteAllSubDomainFromUser(userid string) error

	UpsertCertificate(cert *model.Certificate) error
	GetCertificate(domain string) (*model.Certificate, error)
	DelCertificate(domain string) error

	BlockEmail(email, data string) error
	IsEmailBlocked(email string) (bool, error)

	UpsertUserStatistics(stat *model.UserStatistics) error
	IncrUserStatistics(stat *model.UserStatistics) error
	GetUserStatistics(userid, monthyear string) (*model.UserStatistics, error)
	GetAllUserStatistics(userid string) ([]*model.UserStatistics, error)
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
