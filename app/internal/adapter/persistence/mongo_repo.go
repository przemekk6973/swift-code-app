package persistence

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/port"
)

// MongoRepository implements port.SwiftRepository for MongoDB
type MongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoRepository creates connection with MongoDB and inits collection
func NewMongoRepository(uri, dbName, collName string) (port.SwiftRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	coll := client.Database(dbName).Collection(collName)
	// Indexes
	_, err = coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "swiftCode", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "countryISO2", Value: 1}},
			Options: options.Index().SetBackground(true),
		},
	})
	if err != nil {
		return nil, err
	}

	return &MongoRepository{client: client, collection: coll}, nil

}

// SaveHeadquarters
func (r *MongoRepository) SaveHeadquarters(ctx context.Context, hqs []models.SwiftCode) (models.ImportSummary, error) {
	var summary models.ImportSummary
	for _, hq := range hqs {
		filter := bson.M{"swiftCode": hq.SwiftCode}
		update := bson.M{"$setOnInsert": hq}
		opts := options.Update().SetUpsert(true)
		res, err := r.collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return summary, err
		}
		if res.MatchedCount == 0 {
			summary.HQAdded++
		} else {
			summary.HQSkipped++
		}
	}
	return summary, nil
}

// SaveBranches add branches, checking if HQ exists
func (r *MongoRepository) SaveBranches(ctx context.Context, branches []models.SwiftCode) (models.ImportSummary, error) {
	var summary models.ImportSummary
	for _, br := range branches {
		// get HQ with prefix of 8 characters
		hqCode := strings.ToUpper(br.SwiftCode[:8] + "XXX")
		filter := bson.M{"swiftCode": hqCode}
		// add uniqie
		update := bson.M{"$addToSet": bson.M{"branches": models.SwiftBranch{
			SwiftCode:     br.SwiftCode,
			BankName:      br.BankName,
			Address:       br.Address,
			CountryISO2:   br.CountryISO2,
			IsHeadquarter: false,
		}}}
		res, err := r.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return summary, err
		}
		switch {
		case res.MatchedCount == 0:
			summary.BranchesMissingHQ++
		case res.ModifiedCount == 0:
			summary.BranchesDuplicate++
		default:
			summary.BranchesAdded++
		}
	}
	return summary, nil
}

// GetByCode gets SwiftCode (HQ or branch) by code
func (r *MongoRepository) GetByCode(ctx context.Context, code string) (models.SwiftCode, error) {
	// Fetch the HQ document that either has swiftCode == code OR contains the branch
	filter := bson.M{
		"$or": []bson.M{
			{"swiftCode": code},
			{"branches.swiftCode": code},
		},
	}
	var doc models.SwiftCode
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.SwiftCode{}, port.ErrNotFound
		}
		return models.SwiftCode{}, err
	}

	// If the code matches the HQ itself, return it (with its branches)
	if doc.SwiftCode == code {
		return doc, nil
	}

	// Otherwise it must be one of the branches: find it and return a branch‐only struct
	for _, br := range doc.Branches {
		if br.SwiftCode == code {
			return models.SwiftCode{
				SwiftCode:     br.SwiftCode,
				BankName:      br.BankName,
				Address:       br.Address,
				CountryISO2:   doc.CountryISO2,
				CountryName:   doc.CountryName,
				IsHeadquarter: false,
				// omit Branches slice entirely
			}, nil
		}
	}

	// If somehow it wasn’t in doc.Branches, treat as not found
	return models.SwiftCode{}, port.ErrNotFound
}

// GetByCountry gets all code (HQ and branches) for a country
func (r *MongoRepository) GetByCountry(ctx context.Context, iso2 string) ([]models.SwiftCode, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"countryISO2": iso2})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.SwiftCode
	for cursor.Next(ctx) {
		var hq models.SwiftCode
		if err := cursor.Decode(&hq); err != nil {
			return nil, err
		}
		// add HQ
		results = append(results, hq)
		// add branches
		for _, br := range hq.Branches {
			results = append(results, models.SwiftCode{
				SwiftCode:     br.SwiftCode,
				BankName:      br.BankName,
				Address:       br.Address,
				CountryISO2:   br.CountryISO2,
				CountryName:   hq.CountryName,
				IsHeadquarter: false,
			})
		}
	}
	if len(results) == 0 {
		return nil, port.ErrNotFound
	}
	return results, nil

}

// AddBranch add branch to exisitng HQ
func (r *MongoRepository) AddBranch(ctx context.Context, hqCode string, br models.SwiftBranch) error {
	filter := bson.M{"swiftCode": hqCode}
	update := bson.M{"$addToSet": bson.M{"branches": br}}
	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return port.ErrNotFound
	}
	if res.ModifiedCount == 0 {
		return port.ErrBranchDuplicate
	}
	return nil
}

// Delete deletes entry by given SWIFT code
func (r *MongoRepository) Delete(ctx context.Context, code string) error {
	if strings.HasSuffix(code, "XXX") {
		// usuń cały HQ i oddziały
		res, err := r.collection.DeleteOne(ctx, bson.M{"swiftCode": code})
		if err != nil {
			return err
		}
		if res.DeletedCount == 0 {
			return port.ErrNotFound
		}
		return nil
	}
	// remove branch onlu
	filter := bson.M{"branches.swiftCode": code}
	update := bson.M{"$pull": bson.M{"branches": bson.M{"swiftCode": code}}}
	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return port.ErrNotFound
	}
	return nil
}

func (r *MongoRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, readpref.Primary())
}

// Close closes MongoDB connection
func (r *MongoRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
