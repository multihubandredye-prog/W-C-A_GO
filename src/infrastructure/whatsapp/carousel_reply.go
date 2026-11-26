package whatsapp

import (
	"encoding/json"
	"fmt"

	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

// Carousel IDs 1..10 are reserved by /send/carousel. This identifies our
// carousel buttons even when the client does not return the quoted carousel.
// It is a naming convention, not cryptographic proof of message origin.
func carouselReplyIndex(msg *waE2E.Message) int {
	if msg == nil {
		return 0
	}
	var id string
	if response := msg.GetInteractiveResponseMessage(); response != nil {
		if native := response.GetNativeFlowResponseMessage(); native != nil {
			var data struct {
				ID string `json:"id"`
			}
			if json.Unmarshal([]byte(native.GetParamsJSON()), &data) == nil {
				id = data.ID
			}
		}
	} else if response := msg.GetTemplateButtonReplyMessage(); response != nil {
		id = response.GetSelectedID()
	} else if response := msg.GetButtonsResponseMessage(); response != nil {
		id = response.GetSelectedButtonID()
	}
	for i := 1; i <= 10; i++ {
		if id == fmt.Sprintf("carousel_%d", i) {
			return i
		}
	}
	return 0
}

func buildCarouselReplyFields(msg *waE2E.Message, payload map[string]any) {
	index := carouselReplyIndex(msg)
	if index == 0 {
		return
	}
	reply, ok := payload["InteractiveReply"].(map[string]any)
	if !ok {
		return
	}
	// Keep generic InteractiveReply fields available to existing consumers.
	reply["Type"] = "carousel"
	reply["CardIndex"] = index
	if info := utils.ExtractContextInfo(msg); info != nil && info.GetStanzaID() != "" {
		reply["RepliedToID"] = info.GetStanzaID()
	}
	carousel := make(map[string]any, len(reply))
	for key, value := range reply {
		carousel[key] = value
	}
	payload["CarouselReply"] = carousel
}
