package rest

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
)

func autoCallPolicyGet(c *fiber.Ctx) error {
	d := getDeviceFromCtx(c)
	if d == nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "DEVICE_ID_REQUIRED", Message: "Device not selected"})
	}
	p, e := d.GetAutoCallPolicy()
	if e != nil {
		return c.Status(409).JSON(utils.ResponseData{Status: 409, Code: "POLICY_UNAVAILABLE", Message: "Device or policy is unavailable"})
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Automatic call policy", Results: p})
}
func autoCallPolicyPut(c *fiber.Ctx) error {
	d := getDeviceFromCtx(c)
	if d == nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "DEVICE_ID_REQUIRED", Message: "Device not selected"})
	}
	var p whatsapp.AutoCallPolicy
	if e := c.BodyParser(&p); e != nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Invalid policy JSON"})
	}
	if e := p.Validate(); e != nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: e.Error()})
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(c.Body(), &fields); e != nil || fields == nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Invalid policy JSON"})
	}
	_, hasDelay := fields["answer_delay_seconds"]
	var saveErr error
	if hasDelay {
		saveErr = d.SetAutoCallPolicy(p)
	} else {
		saveErr = d.SetAutoCallPolicyPreservingDelay(p)
	}
	if saveErr != nil {
		return c.Status(500).JSON(utils.ResponseData{Status: 500, Code: "POLICY_SAVE_FAILED", Message: "Could not save automatic call policy"})
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Automatic call policy saved", Results: p})
}

func autoCallContacts(c *fiber.Ctx) error {
	d := getDeviceFromCtx(c)
	if d == nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "DEVICE_ID_REQUIRED", Message: "Device not selected"})
	}
	cli := d.GetClient()
	if cli == nil || cli.Store == nil || cli.Store.Contacts == nil || !cli.IsLoggedIn() {
		return c.Status(409).JSON(utils.ResponseData{Status: 409, Code: "DEVICE_OFFLINE", Message: "Connect the selected device first"})
	}
	contacts, err := cli.Store.Contacts.GetAllContacts(c.UserContext())
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{Status: 500, Code: "CONTACTS_UNAVAILABLE", Message: "Could not load contacts"})
	}
	type item struct {
		Number string `json:"number"`
		Name   string `json:"name"`
	}
	items := make([]item, 0)
	seen := map[string]bool{}
	for jid, contact := range contacts {
		if jid.Server != "s.whatsapp.net" && jid.Server != "lid" {
			continue
		}
		pn := whatsapp.NormalizeJIDFromLID(c.UserContext(), jid, cli).ToNonAD()
		if pn.Server != "s.whatsapp.net" || len(pn.User) < 7 || len(pn.User) > 15 || pn.User[0] == '0' || seen[pn.User] {
			continue
		}
		valid := true
		for _, r := range pn.User {
			if r < '0' || r > '9' {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		seen[pn.User] = true
		items = append(items, item{Number: pn.User, Name: contact.FullName})
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Individual contacts", Results: items})
}

func autoCallCopyAudio(c *fiber.Ctx) error {
	var req struct {
		AudioPath string `json:"audio_path"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Provide an absolute local MP3 audio_path"})
	}
	path, err := whatsapp.CopyAutoCallAudio(req.AudioPath)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "AUDIO_COPY_FAILED", Message: err.Error()})
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Audio copied", Results: map[string]string{"audio_path": path}})
}
