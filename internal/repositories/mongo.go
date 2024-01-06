package repositories

import (
	"context"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBRepository struct {
	collection *mongo.Collection
	modelType  reflect.Type
}

func NewMongoDBRepository(db *mongo.Database, collectionName string, model interface{}) *MongoDBRepository {
	collection := db.Collection(collectionName)
	modelType := reflect.TypeOf(model)
	return &MongoDBRepository{collection: collection, modelType: modelType}
}

func (r *MongoDBRepository) InsertOne(ctx context.Context, model interface{}) (string, error) {
	result, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		return "", err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to assert InsertedID as ObjectID")
	}

	return insertedID.Hex(), nil
}

func (r *MongoDBRepository) FindOne(ctx context.Context, filter interface{}) (interface{}, error) {
	result := reflect.New(r.modelType).Interface()
	err := r.collection.FindOne(ctx, filter).Decode(result)

	return result, err
}

func (r *MongoDBRepository) Find(ctx context.Context, filter interface{}) ([]interface{}, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []interface{}
	for cursor.Next(ctx) {
		result := reflect.New(r.modelType).Interface()
		if err := cursor.Decode(result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *MongoDBRepository) UpdateOne(ctx context.Context, filter, update interface{}) error {
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoDBRepository) UpdateMany(ctx context.Context, filter, update interface{}) error {
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *MongoDBRepository) DeleteOne(ctx context.Context, filter interface{}) error {
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *MongoDBRepository) UpdateOneWithUpsert(ctx context.Context, filter, update interface{}) error {
	opt := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opt)
	return err
}
