package rest

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/multihubandredye-prog/W-C-A_GO/config"
	domainApp "github.com/multihubandredye-prog/W-C-A_GO/domains/app"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/sysmetrics"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
)

type App struct {
	Service domainApp.IAppUsecase
}

// InitRestAppMetrics registers operational metrics outside device middleware.
// The route remains protected by the existing global Basic Auth middleware.
func InitRestAppMetrics(app fiber.Router) {
	app.Get("/app/metrics", AppMetrics)
}

func AppMetrics(c *fiber.Ctx) error {
	snapshot := sysmetrics.Collect(map[string]string{
		"media": config.PathMedia, "storages": config.PathStorages,
	}, config.PathStorages)
	devices := map[string]int{"total": 0, "logged_in": 0, "connected": 0}
	if dm := whatsapp.GetDeviceManager(); dm != nil {
		for _, device := range dm.ListDevices() {
			devices["total"]++
			if device.IsLoggedIn() {
				devices["logged_in"]++
			}
			if device.IsConnected() {
				devices["connected"]++
			}
		}
	}
	sessionStore := "sqlite"
	if strings.HasPrefix(strings.Trim(config.DBURI, "\"'"), "postgres:") {
		sessionStore = "postgres"
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Fetch app metrics success", Results: map[string]any{
		"version": config.AppVersion, "session_store": sessionStore, "chat_storage_persistent": config.ChatStorageEnabled, "devices": devices, "system": snapshot,
	}})
}

func InitRestApp(app fiber.Router, service domainApp.IAppUsecase) App {
	rest := App{Service: service}
	app.Get("/app/login", rest.Login)
	app.Get("/app/login-with-code", rest.LoginWithCode)
	app.Get("/app/logout", rest.Logout)
	app.Get("/app/reconnect", rest.Reconnect)
	app.Get("/app/devices", rest.Devices)
	app.Get("/app/status", rest.ConnectionStatus)
	app.Get("/app/database/check", rest.DatabaseCheck)

	return App{Service: service}
}

// DatabaseCheck handles GET /app/database/check — verifies the integrity of
// the WhatsApp store database (SQLite) so a corrupted file can be detected
// before endpoints such as /user/my/contacts start returning 500 errors.
func (handler *App) DatabaseCheck(c *fiber.Ctx) error {
	result := whatsapp.CheckDatabaseIntegrity(c.UserContext())

	status := 200
	code := "OK"
	if !result.OK {
		status = 500
		code = "CORRUPT"
	}

	return c.Status(status).JSON(utils.ResponseData{
		Status:  status,
		Code:    code,
		Message: "database integrity check",
		Results: result,
	})
}

func (handler *App) Login(c *fiber.Ctx) error {
	device, err := getDeviceInstance(c)
	if err != nil {
		return err
	}

	response, err := handler.Service.Login(c.UserContext(), device.ID())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Login success",
		Results: map[string]any{
			"device_id":   device.ID(),
			"qr_link":     fmt.Sprintf("%s://%s%s/%s", c.Protocol(), c.Hostname(), config.AppBasePath, response.ImagePath),
			"qr_duration": response.Duration,
		},
	})
}

func (handler *App) LoginWithCode(c *fiber.Ctx) error {
	device, err := getDeviceInstance(c)
	if err != nil {
		return err
	}

	pairCode, err := handler.Service.LoginWithCode(c.UserContext(), device.ID(), c.Query("phone"))
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Login with code success",
		Results: map[string]any{
			"device_id": device.ID(),
			"pair_code": pairCode,
		},
	})
}

func (handler *App) Logout(c *fiber.Ctx) error {
	device, err := getDeviceInstance(c)
	if err != nil {
		return err
	}

	err = handler.Service.Logout(c.UserContext(), device.ID())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Success logout",
		Results: map[string]any{"device_id": device.ID()},
	})
}

func (handler *App) Reconnect(c *fiber.Ctx) error {
	device, err := getDeviceInstance(c)
	if err != nil {
		return err
	}

	err = handler.Service.Reconnect(c.UserContext(), device.ID())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Reconnect success",
		Results: map[string]any{"device_id": device.ID()},
	})
}

func (handler *App) Devices(c *fiber.Ctx) error {
	devices, err := handler.Service.FetchDevices(c.UserContext())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Fetch device success",
		Results: devices,
	})
}

func (handler *App) ConnectionStatus(c *fiber.Ctx) error {
	device, err := getDeviceInstance(c)
	if err != nil {
		return err
	}

	isConnected, isLoggedIn, err := handler.Service.Status(c.UserContext(), device.ID())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Connection status retrieved",
		Results: map[string]any{
			"is_connected": isConnected,
			"is_logged_in": isLoggedIn,
			"device_id":    device.ID(),
		},
	})
}

func getDeviceInstance(c *fiber.Ctx) (*whatsapp.DeviceInstance, error) {
	value := c.Locals("device")
	if value == nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "device context is missing")
	}
	device, ok := value.(*whatsapp.DeviceInstance)
	if !ok || device == nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid device context")
	}
	return device, nil
}
