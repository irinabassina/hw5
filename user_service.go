package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"strconv"
	"time"
)

type userService struct {
	dbName string
}

func newUserService(dbName string) *userService {
	db := openDB(dbName)
	defer db.Close()
	return &userService{dbName: dbName}
}

func openDB(dbName string) *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(dbName))
	if err != nil {
		panic(err)
	}
	return db
}

func encodeUser(user *User) []byte {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err := e.Encode(user); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func decodeUser(value []byte) *User {
	var user User
	d := gob.NewDecoder(bytes.NewReader(value))
	if err := d.Decode(&user); err != nil {
		panic(err)
	}
	return &user
}

func getUserFromDB(id string, db *badger.DB) *User {
	var user *User
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		user = decodeUser(val)
		return nil
	})

	if err != nil && err.Error() != "Key not found" {
		fmt.Printf("ERROR reading from badger db : %s\n", err)
	}
	return user
}

func storeUserToDB(user *User, db *badger.DB) error {
	if user.ID == "" {
		user.ID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(user.ID), encodeUser(user))
		return err
	})
	if err != nil {
		fmt.Printf("ERROR saving to badger db : %s\n", err)
		return err
	}
	return nil
}

func (us *userService) getUser(id string) *User {
	db := openDB(us.dbName)
	defer db.Close()
	return getUserFromDB(id, db)
}

func (us *userService) storeUser(user *User) string {
	db := openDB(us.dbName)
	defer db.Close()
	err := storeUserToDB(user, db)
	if err != nil {
		return ""
	}
	return user.ID
}

func (us *userService) deleteUser(targetID string) (string, bool) {
	db := openDB(us.dbName)
	defer db.Close()

	userToDelete := getUserFromDB(targetID, db)

	if userToDelete != nil {
		for _, friendID := range userToDelete.Friends {
			us.deleteFromFriends(friendID, userToDelete.ID, db)
		}

		err := db.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(targetID))
		})
		if err != nil {
			fmt.Printf("ERROR deleting from badger db : %s\n", err)
		}

		return userToDelete.Name, true
	}
	return "", false
}

func (us *userService) deleteFromFriends(srcUserID string, toDeleteID string, db *badger.DB) {
	u := getUserFromDB(srcUserID, db)
	deleteIdx := -1
	for i, userID := range u.Friends {
		if userID == toDeleteID {
			deleteIdx = i
			break
		}
	}
	if deleteIdx != -1 {
		u.Friends = append(u.Friends[:deleteIdx], u.Friends[deleteIdx+1:]...)
	}

	storeUserToDB(u, db)
}

func (us *userService) makeFriends(friends *Friends) (string, string, bool) {
	db := openDB(us.dbName)
	defer db.Close()

	srcUser := getUserFromDB(friends.SourceID, db)
	targetUser := getUserFromDB(friends.TargetID, db)
	if srcUser == nil || targetUser == nil {
		return "", "", false
	}

	for _, friendID := range srcUser.Friends {
		if friendID == targetUser.ID {
			return srcUser.Name, targetUser.Name, true
		}
	}

	srcUser.Friends = append(srcUser.Friends, friends.TargetID)
	targetUser.Friends = append(targetUser.Friends, friends.SourceID)
	storeUserToDB(srcUser, db)
	storeUserToDB(targetUser, db)
	return srcUser.Name, targetUser.Name, true
}

func (us *userService) getFriends(userID string) ([]*User, bool) {
	db := openDB(us.dbName)
	defer db.Close()

	user := getUserFromDB(userID, db)
	if user == nil {
		return nil, false
	}
	var friends []*User
	for _, friendID := range user.Friends {
		friends = append(friends, getUserFromDB(friendID, db))
	}
	return friends, true
}

func (us *userService) updateAge(userID string, age string) bool {
	db := openDB(us.dbName)
	defer db.Close()

	user := getUserFromDB(userID, db)
	if user == nil {
		return false
	}
	user.Age = age
	storeUserToDB(user, db)
	return true
}

func (us *userService) dropDB() {
	db := openDB(us.dbName)
	db.DropAll()
}
