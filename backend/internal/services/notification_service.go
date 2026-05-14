package services

import (
	"encoding/json"
	"log"
	"sync"

	"fly-box/backend/internal/models"
	"fly-box/backend/internal/repository"
)

// NotificationService handles notification events and broadcasting.
type NotificationService struct {
	repo   *repository.Repository
	hub    interface {
		BroadcastUser(userID uint, event interface{})
	}
	listeners map[uint][]chan<- models.NotificationEvent
	listenerMu sync.RWMutex
}

var notifServiceInstance *NotificationService
var notifServiceOnce sync.Once

// NewNotificationService creates a singleton NotificationService.
func NewNotificationService(repo *repository.Repository, hub interface {
	BroadcastUser(userID uint, event interface{})
}) *NotificationService {
	notifServiceOnce.Do(func() {
		notifServiceInstance = &NotificationService{
			repo:      repo,
			hub:       hub,
			listeners: make(map[uint][]chan<- models.NotificationEvent),
		}
	})
	return notifServiceInstance
}

// EmitNotification creates a notification and broadcasts it to the user.
func (s *NotificationService) EmitNotification(event models.NotificationEvent) {
	// Convert event data to JSON
	var dataJSON []byte
	if event.Data != nil {
		dataJSON, _ = json.Marshal(event.Data)
	}

	notification := &models.Notification{
		UserID:   event.UserID,
		Title:    event.Title,
		Body:     event.Body,
		Type:     event.Type,
		Platform: event.Platform,
		IsRead:   false,
	}

	if event.PageID != 0 {
		pageID := event.PageID
		notification.PageID = &pageID
	}

	if len(dataJSON) > 0 {
		notification.Data = dataJSON
	}

	// Save to database
	if err := s.repo.CreateNotification(notification); err != nil {
		log.Printf("[notification] failed to save: %v", err)
		return
	}

	// Broadcast to WebSocket
	s.hub.BroadcastUser(event.UserID, map[string]interface{}{
		"type":         "NEW_NOTIFICATION",
		"notification": notification,
	})

	// Notify listeners
	s.notifyListeners(event.UserID, event)
}

// notifyListeners sends the event to all registered listeners for a user.
func (s *NotificationService) notifyListeners(userID uint, event models.NotificationEvent) {
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()

	chans, ok := s.listeners[userID]
	if !ok {
		return
	}

	for _, ch := range chans {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// Subscribe returns a channel that receives notification events for a user.
func (s *NotificationService) Subscribe(userID uint) chan models.NotificationEvent {
	ch := make(chan models.NotificationEvent, 100)

	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()

	s.listeners[userID] = append(s.listeners[userID], ch)
	return ch
}

// Unsubscribe removes a listener channel.
func (s *NotificationService) Unsubscribe(userID uint, ch chan models.NotificationEvent) {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()

	chans := s.listeners[userID]
	for i, c := range chans {
		if c == ch {
			s.listeners[userID] = append(chans[:i], chans[i+1:]...)
			close(ch)
			break
		}
	}
}

// EmitNewMessageNotification creates a notification for a new message.
func (s *NotificationService) EmitNewMessageNotification(userID, pageID, convID uint, platform, customerName, messagePreview string) {
	event := models.NotificationEvent{
		Type:     models.NotificationTypeNewMessage,
		Platform: platform,
		UserID:   userID,
		PageID:   pageID,
		ConvID:   convID,
		Title:    "Tin nhắn mới",
		Body:     customerName + ": " + messagePreview,
		Data: map[string]interface{}{
			"conversation_id":  convID,
			"customer_name":    customerName,
			"message_preview":  messagePreview,
		},
	}
	s.EmitNotification(event)
}

// EmitNewOrderNotification creates a notification for a new order.
func (s *NotificationService) EmitNewOrderNotification(userID, pageID uint, platform, orderInfo string) {
	event := models.NotificationEvent{
		Type:     models.NotificationTypeNewOrder,
		Platform: platform,
		UserID:   userID,
		PageID:   pageID,
		Title:    "Đơn hàng mới",
		Body:     orderInfo,
	}
	s.EmitNotification(event)
}

// EmitNewFollowerNotification creates a notification for a new follower.
func (s *NotificationService) EmitNewFollowerNotification(userID, pageID uint, platform, followerName string) {
	event := models.NotificationEvent{
		Type:     models.NotificationTypeNewFollower,
		Platform: platform,
		UserID:   userID,
		PageID:   pageID,
		Title:    "Ng��ời theo dõi mới",
		Body:     followerName + " đã theo dõi bạn",
	}
	s.EmitNotification(event)
}

// EmitSystemNotification creates a system notification.
func (s *NotificationService) EmitSystemNotification(userID uint, title, body string) {
	event := models.NotificationEvent{
		Type:   models.NotificationTypeSystem,
		UserID: userID,
		Title:  title,
		Body:   body,
	}
	s.EmitNotification(event)
}