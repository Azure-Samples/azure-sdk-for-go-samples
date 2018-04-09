// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmos

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetCollection gets a mongoDB collection
func GetCollection(session *mgo.Session, database, collectionName string) (*mgo.Collection, error) {
	collection := session.DB(database).C(collectionName)
	return collection, nil
}

// InsertDocument inserts a mongoDB document in the specified collection
func InsertDocument(session *mgo.Session, database, collectionName string, document map[string]interface{}) error {
	collection, err := GetCollection(session, database, collectionName)
	if err != nil {
		return err
	}
	return collection.Insert(document)
}

// GetDocument gets a mongoDB document in the specified collection
func GetDocument(session *mgo.Session, database, collectionName string, query bson.M) (result map[string]interface{}, err error) {
	collection, err := GetCollection(session, database, collectionName)
	if err != nil {
		return
	}

	err = collection.Find(query).One(&result)
	return
}

// UpdateDocument updates the mongoDB document with the specified ID
func UpdateDocument(session *mgo.Session, database, collectionName string, id bson.ObjectId, change bson.M) error {
	collection, err := GetCollection(session, database, collectionName)
	if err != nil {
		return err
	}
	return collection.Update(bson.M{"_id": id}, change)
}

// DeleteDcoument deletes the mongoDB document with the specified ID
func DeleteDcoument(session *mgo.Session, database, collectionName string, id bson.ObjectId) error {
	collection, err := GetCollection(session, database, collectionName)
	if err != nil {
		return err
	}
	return collection.Remove(bson.M{"_id": id})
}
