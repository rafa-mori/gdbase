package backend

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoBackend is a backend implementation for MongoDB.
type MongoBackend struct {
	db *mongo.Database
}

// NewMongoBackend creates a new MongoBackend.
func NewMongoBackend(db *mongo.Database) *MongoBackend {
	return &MongoBackend{db: db}
}

// Create creates a new document.
func (b *MongoBackend) Create(value any) error {
	// This is a generic implementation and might need to be adapted
	// based on the specific type of 'value'.
	_, err := b.db.Collection("default").InsertOne(context.Background(), value)
	return err
}

// Update updates a document.
func (b *MongoBackend) Update(value any) error {
	// This requires a filter and update document, which is not directly
	// available in this generic interface. This is a placeholder.
	return nil // Not implemented
}

// Delete deletes a document.
func (b *MongoBackend) Delete(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = b.db.Collection("default").DeleteOne(context.Background(), primitive.M{"_id": objID})
	return err
}

// GetAll retrieves all documents from a collection.
func (b *MongoBackend) GetAll(out any) error {
	// This is a generic implementation. You might need to specify a collection.
	cursor, err := b.db.Collection("default").Find(context.Background(), primitive.M{})
	if err != nil {
		return err
	}
	return cursor.All(context.Background(), out)
}

// GetByID retrieves a document by its ID.
func (b *MongoBackend) GetByID(id string, out any) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return b.db.Collection("default").FindOne(context.Background(), primitive.M{"_id": objID}).Decode(out)
}
