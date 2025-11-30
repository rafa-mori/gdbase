package infra

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo[T any] struct {
	DB      *mongo.Database
	Coll    *mongo.Collection
	Timeout time.Duration
}

// NewMongoRepo cria um repo simples com timeout default (8s).
func NewMongoRepo[T any](db *mongo.Database, collName string) *MongoRepo[T] {
	if db == nil {
		panic("NewMongoRepo: db nil")
	}
	return &MongoRepo[T]{
		DB:      db,
		Coll:    db.Collection(collName),
		Timeout: 8 * time.Second,
	}
}

func (r *MongoRepo[T]) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if r.Timeout <= 0 {
		return context.WithTimeout(ctx, 8*time.Second)
	}
	return context.WithTimeout(ctx, r.Timeout)
}

func (r *MongoRepo[T]) InsertOne(ctx context.Context, doc *T) (primitive.ObjectID, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	res, err := r.Coll.InsertOne(ctx, doc)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return oid, nil
	}
	// se o documento já traz _id custom (string, etc.), ignore conversão
	return primitive.NilObjectID, nil
}

func (r *MongoRepo[T]) FindByID(ctx context.Context, id primitive.ObjectID) (*T, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	var out T
	err := r.Coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *MongoRepo[T]) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*T, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	var out T
	err := r.Coll.FindOne(ctx, filter, opts...).Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *MongoRepo[T]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (int64, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	res, err := r.Coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (r *MongoRepo[T]) DeleteOne(ctx context.Context, filter interface{}) (int64, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	res, err := r.Coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

// EnsureIndexes cria índices simples idempotentes.
func (r *MongoRepo[T]) EnsureIndexes(ctx context.Context, models []mongo.IndexModel) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	_, err := r.Coll.Indexes().CreateMany(ctx, models)
	return err
}
