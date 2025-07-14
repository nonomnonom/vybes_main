package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BookmarkRepository defines the interface for bookmark data operations.
type BookmarkRepository interface {
	Add(ctx context.Context, bookmark *domain.Bookmark) error
	Remove(ctx context.Context, userID, postID primitive.ObjectID) error
	FindByUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]domain.Bookmark, error)
}

type mongoBookmarkRepository struct {
	db *mongo.Database
}

// NewMongoBookmarkRepository creates a new bookmark repository.
func NewMongoBookmarkRepository(db *mongo.Database) BookmarkRepository {
	return &mongoBookmarkRepository{db: db}
}

func (r *mongoBookmarkRepository) bookmarks() *mongo.Collection {
	return r.db.Collection("bookmarks")
}

func (r *mongoBookmarkRepository) Add(ctx context.Context, bookmark *domain.Bookmark) error {
	// Use upsert to avoid duplicate bookmarks
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"userid": bookmark.UserID, "postid": bookmark.PostID}
	update := bson.M{"$setOnInsert": bookmark}
	_, err := r.bookmarks().UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *mongoBookmarkRepository) Remove(ctx context.Context, userID, postID primitive.ObjectID) error {
	_, err := r.bookmarks().DeleteOne(ctx, bson.M{"userid": userID, "postid": postID})
	return err
}

func (r *mongoBookmarkRepository) FindByUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]domain.Bookmark, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}).SetLimit(int64(limit))
	filter := bson.M{"userid": userID}
	cursor, err := r.bookmarks().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookmarks []domain.Bookmark
	if err = cursor.All(ctx, &bookmarks); err != nil {
		return nil, err
	}
	return bookmarks, nil
}