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
type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetNotificationsForUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, notificationIDs []primitive.ObjectID, userID primitive.ObjectID) (int64, error)
}

type mongoNotificationRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoNotificationRepository creates a new notification repository.
func NewMongoNotificationRepository(db *mongo.Database) NotificationRepository {
	return &mongoNotificationRepository{
		db:         db,
		collection: "notifications",
	}
}

func (r *mongoNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	_, err := r.db.Collection(r.collection).InsertOne(ctx, notification)
	return err
}

func (r *mongoNotificationRepository) GetNotificationsForUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]domain.Notification, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}).SetLimit(int64(limit))
	filter := bson.M{"userid": userID}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
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

func (r *mongoNotificationRepository) MarkAsRead(ctx context.Context, notificationIDs []primitive.ObjectID, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"_id":    bson.M{"$in": notificationIDs},
		"userid": userID, // Ensure users can only mark their own notifications as read
	}
	update := bson.M{"$set": bson.M{"read": true}}
	result, err := r.db.Collection(r.collection).UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}