package persistence

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/port"
)

// MongoRepository implementuje port.SwiftRepository dla MongoDB
type MongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoRepository tworzy nowe połączenie z MongoDB i inicjuje kolekcję
func NewMongoRepository(uri, dbName, collName string) (port.SwiftRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	coll := client.Database(dbName).Collection(collName)
	// Indeksy
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

// SaveHeadquarters implementacja: upsert dokumentów HQ
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

// SaveBranches implementacja: dodajemy oddziały, sprawdzając istnienie HQ
func (r *MongoRepository) SaveBranches(ctx context.Context, branches []models.SwiftCode) (models.ImportSummary, error) {
	var summary models.ImportSummary
	for _, br := range branches {
		// wylicz kod HQ po prefiksie 8 znaków
		hqCode := strings.ToUpper(br.SwiftCode[:8] + "XXX")
		filter := bson.M{"swiftCode": hqCode}
		// dodaj do tablicy branches unikalnie
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

// GetByCode pobiera SwiftCode (HQ lub oddział) po kodzie
func (r *MongoRepository) GetByCode(ctx context.Context, code string) (models.SwiftCode, error) {
	// 1) Fetch the HQ document that either has swiftCode == code OR contains the branch
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

	// 2) If the code matches the HQ itself, return it (with its branches)
	if doc.SwiftCode == code {
		return doc, nil
	}

	// 3) Otherwise it must be one of the branches: find it and return a branch‐only struct
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

	// 4) If somehow it wasn’t in doc.Branches, treat as not found
	return models.SwiftCode{}, port.ErrNotFound
}

// GetByCountry pobiera wszystkie kody (HQ i oddziały) dla danego kraju
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
		// dodaj HQ
		results = append(results, hq)
		// dodaj oddziały
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

// AddBranch dodaje oddział do istniejącego HQ
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

// Delete usuwa wpis po podanym kodzie SWIFT
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
	// usuń tylko oddział
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
