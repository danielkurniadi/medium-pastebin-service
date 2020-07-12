package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/iqdf/pastebin-service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// CollectionName is name of collection
const CollectionName = "paste"

// PasteDBModel ...
type PasteDBModel struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	Title        string    `bson:"title"`
	AuthorUserID string    `bson:"authorUserID,omitempty"`
	ShortURLPath string    `bson:"shortURLPath,omitempty"`
	TextData     string    `bson:"textData,omitempty"`
	StorageURL   string    `bson:"storageUrl,omitempty"`
	Private      bool      `bson:"private"`
	CreatedAt    time.Time `bson:"createdAt,omitempty"`
	UpdatedAt    time.Time `bson:"updatedAt,omitempty"`
	ExpiredAt    time.Time `bson:"expiredAt,omitempty"`
}

// PasteMongoRepo represent mongoDB database
// and handles Paste resources
type PasteMongoRepo struct {
	client *mongo.Client
	db     *mongo.Database
}

// Collection returns mongoDB collection for
// managing paste resources
func (repo *PasteMongoRepo) Collection() *mongo.Collection {
	return repo.db.Collection(CollectionName)
}

func modelFromEntity(paste domain.Paste) PasteDBModel {
	return PasteDBModel{
		Title:        paste.Title,
		AuthorUserID: paste.AuthorUserID,
		ShortURLPath: paste.ShortURLPath,
		TextData:     paste.TextData,
		StorageURL:   paste.StorageURL,
		Private:      paste.Private,
		UpdatedAt:    time.Now(),
	}
}

// Paste creates entity/domain (data transfer object)
// from model field attributes.
func (model PasteDBModel) Paste() domain.Paste {
	return domain.Paste{
		Title:        model.Title,
		AuthorUserID: model.AuthorUserID,
		ShortURLPath: model.ShortURLPath,
		TextData:     model.TextData,
		Private:      model.Private,
		StorageURL:   model.StorageURL,
		ExpiredAt:    model.ExpiredAt,
	}
}

// NewPasteRepo creates repository using connection to mongo database
func NewPasteRepo(client *mongo.Client, dbName string) *PasteMongoRepo {
	var (
		db         = client.Database(dbName)
		collection = db.Collection(CollectionName)
	)

	dbIndices := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "expiredAt", Value: bsonx.Int32(1)}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
		{
			Keys:    bsonx.Doc{{Key: "shortURLPath", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},
	}

 	collection.Indexes().CreateMany(
		context.Background(),
		dbIndices,
	)

	return &PasteMongoRepo{
		client: client,
		db:     db,
	}
}

// GetPaste finds and retrieves a single paste data given paste shortURLPath,
// For example, paste with url: https://fakepastebin.org/oiua21jhqw1
// will have short url: oiua21jhqw1
// Error will be returned if no paste found with the given URL path.
func (repo *PasteMongoRepo) GetPaste(ctx context.Context, shortURLPath string) (domain.Paste, error) {
	var model PasteDBModel

	collection := repo.Collection()
	filter := &bson.M{"shortURLPath": shortURLPath}
	err := collection.FindOne(ctx, filter).Decode(&model)

	if err != nil {
		return domain.Paste{}, err
	}
	return model.Paste(), nil
}

// CreatePaste stores new paste content in database
// Error will be returned if invalid / conflicting fields are found or
// databases failed to create new entry.
func (repo *PasteMongoRepo) CreatePaste(ctx context.Context, paste domain.Paste) error {
	var model = modelFromEntity(paste)
	model.CreatedAt = time.Now()

	collection := repo.Collection()
	_, err := collection.InsertOne(ctx, model)

	return err
}

// DeletePaste removes paste from database
// Error will be returned if deletion failed due to
// paste not found, conflicting relation key or database failures.
func (repo *PasteMongoRepo) DeletePaste(ctx context.Context, shortURLPath string) error {
	collection := repo.Collection()
	filter := PasteDBModel{ShortURLPath: shortURLPath}
	_, err := collection.DeleteOne(ctx, filter)

	return err
}
