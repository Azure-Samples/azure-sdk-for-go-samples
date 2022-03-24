// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmosdb

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetCollection gets a mongoDB collection
func GetCollection(session *mgo.Session, database, collectionName string) *mgo.Collection {
	collection := session.DB(database).C(collectionName)
	return collection
}

// InsertDocument inserts a mongoDB document in the specified collection
func InsertDocument(session *mgo.Session, database, collectionName string, document map[string]interface{}) error {
	collection := GetCollection(session, database, collectionName)
	return collection.Insert(document)
}

// GetDocument gets a mongoDB document in the specified collection
func GetDocument(session *mgo.Session, database, collectionName string, query bson.M) (result map[string]interface{}, err error) {
	collection := GetCollection(session, database, collectionName)
	err = collection.Find(query).One(&result)
	return
}

// UpdateDocument updates the mongoDB document with the specified ID
func UpdateDocument(session *mgo.Session, database, collectionName string, id bson.ObjectId, change bson.M) error {
	collection := GetCollection(session, database, collectionName)
	return collection.Update(bson.M{"_id": id}, change)
}

// DeleteDocument deletes the mongoDB document with the specified ID
func DeleteDocument(session *mgo.Session, database, collectionName string, id bson.ObjectId) error {
	collection := GetCollection(session, database, collectionName)
	return collection.Remove(bson.M{"_id": id})
}
