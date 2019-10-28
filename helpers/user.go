package helpers

import (
	"io/ioutil"

	"github.com/TankerHQ/identity-go/identity"
	uuid "github.com/satori/go.uuid"
)

type User struct {
	AppID          string
	Url            string
	UserID         string
	Identity       string
	PublicIdentity string
}

func (app App) CreateUser() User {
	userID := uuid.NewV4().String()
	userIdentity, _ := identity.Create(app.IdConfig, userID)
	publicIdentity, _ := identity.GetPublicIdentity(*userIdentity)
	return User{
		AppID:          app.Descriptor.ID,
		Url:            app.Config.URL,
		UserID:         userID,
		Identity:       *userIdentity,
		PublicIdentity: *publicIdentity,
	}
}

func (user User) CreateDevice() (*Device, error) {
	dir, err := ioutil.TempDir("", user.UserID+"-")
	if err != nil {
		return nil, err
	}
	return &Device{user, dir}, nil
}
