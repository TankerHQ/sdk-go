package helpers

import (
	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
)

type App struct {
	AdminSession *core.Admin
	Descriptor   *core.AppDescriptor
	Config       ServerConfig
	OidcConfig   OidcConfig
	IdConfig     identity.Config
}

func (app *App) GetVerificationCode(email string) (*string, error) {
	return app.AdminSession.GetVerificationCode(app.Descriptor.ID, email)
}

func (app App) EnableOidc() error {
	return app.AdminSession.Update(app.Descriptor.ID, app.OidcConfig.ClientId, app.OidcConfig.Provider)
}

func (app *App) Destroy() error {
	err := app.AdminSession.DeleteApp(app.Descriptor.ID)
	if err != nil {
		return err
	}
	app.AdminSession.Destroy()
	return nil
}
