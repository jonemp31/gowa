package whatsapp

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func handleMessage(ctx context.Context, evt *events.Message, chatStorageRepo domainChatStorage.IChatStorageRepository, client *whatsmeow.Client) {
	// Log message metadata
	metaParts := buildMessageMetaParts(evt)
	log.Infof("Received message %s from %s (%s): %+v",
		evt.Info.ID,
		evt.Info.SourceString(),
		strings.Join(metaParts, ", "),
		evt.Message,
	)

	if err := chatStorageRepo.CreateMessage(ctx, evt); err != nil {
		// Log storage errors to avoid silent failures that could lead to data loss
		log.Errorf("Failed to store incoming message %s: %v", evt.Info.ID, err)
	}

	// Handle image message if present
	handleImageMessage(ctx, evt, client)

	// Auto-mark message as read if configured
	handleAutoMarkRead(ctx, evt, client)

	// Handle auto-reply if configured
	handleAutoReply(ctx, evt, chatStorageRepo, client)

	// Forward to webhook if configured
	handleWebhookForward(ctx, evt, client)
}

func buildMessageMetaParts(evt *events.Message) []string {
	metaParts := []string{
		fmt.Sprintf("pushname: %s", evt.Info.PushName),
		fmt.Sprintf("timestamp: %s", evt.Info.Timestamp),
	}
	if evt.Info.Type != "" {
		metaParts = append(metaParts, fmt.Sprintf("type: %s", evt.Info.Type))
	}
	if evt.Info.Category != "" {
		metaParts = append(metaParts, fmt.Sprintf("category: %s", evt.Info.Category))
	}
	if evt.IsViewOnce {
		metaParts = append(metaParts, "view once")
	}
	return metaParts
}

func handleImageMessage(ctx context.Context, evt *events.Message, client *whatsmeow.Client) {
	if !config.WhatsappAutoDownloadMedia {
		return
	}
	if client == nil {
		return
	}
	if img := evt.Message.GetImageMessage(); img != nil {
		if extracted, err := utils.ExtractMedia(ctx, client, config.PathStorages, img); err != nil {
			log.Errorf("Failed to download image: %v", err)
		} else {
			log.Infof("Image downloaded to %s", extracted.MediaPath)
		}
	}
}

// getMarkReadDelay returns a random delay based on the current hour (America/Sao_Paulo).
// Simulates human reading behavior: faster during active hours, slower at night.
func getMarkReadDelay() time.Duration {
	hour := time.Now().Hour() // TZ=America/Sao_Paulo set in container
	var minSec, maxSec int
	switch {
	case hour >= 8 && hour < 11: // 08:01–11:00
		minSec, maxSec = 6, 18
	case hour >= 11 && hour < 13: // 11:01–13:00
		minSec, maxSec = 18, 36
	case hour >= 13 && hour < 18: // 13:01–18:00
		minSec, maxSec = 8, 24
	case hour >= 18 && hour < 19: // 18:01–19:00
		minSec, maxSec = 20, 40
	case hour >= 19 && hour <= 23: // 19:01–00:00
		minSec, maxSec = 7, 18
	case hour == 0: // 00:01–01:00
		minSec, maxSec = 7, 18
	case hour >= 1 && hour < 2: // 00:01–02:00
		minSec, maxSec = 18, 48
	case hour >= 2 && hour < 6: // 02:01–06:00
		minSec, maxSec = 37, 90
	case hour >= 6 && hour < 8: // 06:01–08:00
		minSec, maxSec = 15, 30
	default:
		minSec, maxSec = 6, 18
	}
	return time.Duration(minSec+rand.Intn(maxSec-minSec+1)) * time.Second
}

func handleAutoMarkRead(ctx context.Context, evt *events.Message, client *whatsmeow.Client) {
	// Only mark read if auto-mark read is enabled and message is incoming
	if !config.WhatsappAutoMarkRead || evt.Info.IsFromMe {
		return
	}

	if client == nil {
		return
	}

	// Capture values for the goroutine
	messageIDs := []types.MessageID{evt.Info.ID}
	chat := evt.Info.Chat
	sender := evt.Info.Sender

	// Mark as read after a random human-like delay (async, non-blocking)
	go func() {
		delay := getMarkReadDelay()
		log.Debugf("Will mark message %s as read in %v", evt.Info.ID, delay)
		time.Sleep(delay)
		markCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.MarkRead(markCtx, messageIDs, time.Now(), chat, sender); err != nil {
			log.Warnf("Failed to mark message %s as read: %v", evt.Info.ID, err)
		} else {
			log.Debugf("Marked message %s as read after %v delay", evt.Info.ID, delay)
		}
	}()
}

func handleWebhookForward(ctx context.Context, evt *events.Message, client *whatsmeow.Client) {
	// Skip webhook for protocol messages that are internal sync messages
	if protocolMessage := evt.Message.GetProtocolMessage(); protocolMessage != nil {
		protocolType := protocolMessage.GetType().String()
		// Only allow REVOKE and MESSAGE_EDIT through - skip all other protocol messages
		// (HISTORY_SYNC_NOTIFICATION, APP_STATE_SYNC_KEY_SHARE, EPHEMERAL_SYNC_RESPONSE, etc.)
		switch protocolType {
		case "REVOKE", "MESSAGE_EDIT":
			// These are meaningful user actions, allow webhook
		default:
			log.Debugf("Skipping webhook for protocol message type: %s", protocolType)
			return
		}
	}

	if (len(config.WhatsappWebhook) > 0 || config.ChatwootEnabled) &&
		!strings.Contains(evt.Info.SourceString(), "broadcast") {
		go func(e *events.Message, c *whatsmeow.Client) {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := forwardMessageToWebhook(webhookCtx, c, e); err != nil {
				logrus.Error("Failed forward to webhook: ", err)
			}
		}(evt, client)
	}
}
