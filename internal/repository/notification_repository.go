package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NotificationRepository defines the interface for notification data operations.
// Notifications inform users about various events such as new followers,
// likes, comments, and other social interactions.
type NotificationRepository interface {
	// CreateNotification creates a new notification for a user
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	// GetUserNotifications retrieves notifications for a specific user
	GetUserNotifications(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Notification, error)
	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, notificationID, userID primitive.ObjectID) error
	// MarkAllAsRead marks all notifications for a user as read
	MarkAllAsRead(ctx context.Context, userID primitive.ObjectID) error
	// GetUnreadCount returns the number of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID primitive.ObjectID) (int64, error)
	// DeleteNotification removes a notification from the database
	DeleteNotification(ctx context.Context, notificationID, userID primitive.ObjectID) error
}

// mongoNotificationRepository implements NotificationRepository using MongoDB as the backend
type mongoNotificationRepository struct {
	collection *mongo.Collection
}

// NewMongoNotificationRepository creates a new notification repository instance with MongoDB backend.
// The repository handles all notification-related database operations including
// creating, reading, updating, and deleting user notifications.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - NotificationRepository: A configured notification repository ready for use
func NewMongoNotificationRepository(db *mongo.Database) NotificationRepository {
	return &mongoNotificationRepository{
		collection: db.Collection("notifications"),
	}
}

// CreateNotification adds a new notification to the database for a specific user.
// The notification contains information about the event that triggered it.
//
// Parameters:
//   - ctx: Context for the operation
//   - notification: The notification object to create
//
// Returns:
//   - error: Any error that occurred during the operation
func (r *mongoNotificationRepository) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	_, err := r.collection.InsertOne(ctx, notification)
	return err
}

// GetUserNotifications retrieves paginated notifications for a specific user.
// Notifications are sorted by creation date in descending order (newest first).
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user whose notifications to retrieve
//   - page: Page number for pagination (0-based)
//   - limit: Maximum number of notifications to return
//
// Returns:
//   - []domain.Notification: List of notifications for the user
//   - error: Any error that occurred during the operation
func (r *mongoNotificationRepository) GetUserNotifications(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Notification, error) {
	skip := int64(page * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "createdat", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))
	
	filter := bson.M{"userid": userID}
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []domain.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkAsRead marks a specific notification as read by updating its read status.
// Only the notification owner can mark their own notifications as read.
//
// Parameters:
//   - ctx: Context for the operation
//   - notificationID: ID of the notification to mark as read
//   - userID: ID of the user who owns the notification
//
// Returns:
//   - error: Any error that occurred during the operation
func (r *mongoNotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID primitive.ObjectID) error {
	filter := bson.M{
		"_id":    notificationID,
		"userid": userID, // Ensure users can only mark their own notifications as read
	}
	update := bson.M{"$set": bson.M{"read": true}}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
