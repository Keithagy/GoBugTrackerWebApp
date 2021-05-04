package dsa

import (
	"fmt"
)

const (
	Hashbuckets = 50 // Length of index array in hash table
)

// UserNode specifies a SLL node in the Userlog hash table.
type UserNode struct {
	User User
	next *UserNode
}

// User struct logs fields for managing login user/admin accounts.
type User struct {
	Name  string
	Pw    []byte
	Admin bool
}

// EmptyUser is a placeholder variable for functions to return a nil result.
var EmptyUser = User{
	Name:  "",
	Pw:    []byte{},
	Admin: false,
}

// Hash table operations

// NewHT initializes the hash table.
func NewHT() *[]*UserNode {
	var Userlog = make([]*UserNode, Hashbuckets)
	return &Userlog
}

// PrintHT returns a slice of slices of strings to be passed into the relevant HTML template for printing in the client. Reflects all users currently recorded in hash table.
func PrintHT(hashtable *[]*UserNode, printfunc func(SLL *UserNode) []string) [][]string {
	empty := true
	str := [][]string{}
	for i := 0; i < Hashbuckets; i++ {
		if (*hashtable)[i] != nil {
			empty = false
			if ls := printfunc((*hashtable)[i]); len(ls) > 0 {
				str = append(str, ls)
			}
		}
	}

	if empty {
		fmt.Println("No logged users!")
	}
	return str
}

// AddUser adds UserNode to SLL (hash bucket). Appends to front of the SLL for O(1) addition.
func AddUser(hashtable *[]*UserNode, newuser User) {
	hashindex := hash(newuser.Name)
	hashbucket := (*hashtable)[hashindex]

	if hashbucket == nil {
		(*hashtable)[hashindex] = &UserNode{newuser, nil}
	} else {
		newnode := &UserNode{newuser, hashbucket}
		(*hashtable)[hashindex] = newnode
	}
}

// EditUser modifies an existing user.
func EditUser(hashtable *[]*UserNode, retrieved, edited User) {
	DeleteUser(hashtable, retrieved.Name)
	AddUser(hashtable, edited)
}

// DeleteUser deletes UserNode with corresponding username from SLL(hash bucket). Assumes username is already verified to exist.
func DeleteUser(hashtable *[]*UserNode, username string) {
	hashindex := hash(username)

	var prev *UserNode
	for ptr := (*hashtable)[hashindex]; ptr != nil; ptr = ptr.next {
		if ptr.User.Name == username {
			if ptr == (*hashtable)[hashindex] { // If first node of SLL
				(*hashtable)[hashindex] = nil
			} else {
				prev.next = ptr.next
			}
		}
		prev = ptr
		fmt.Println(username, "deleted from user log.")
	}
}

// SearchUser looks up particular value from hash table: checks for presence of a particular username.
func SearchUser(hashtable *[]*UserNode, username string) (bool, *UserNode) {
	hashindex := hash(username)

	for ptr := (*hashtable)[hashindex]; ptr != nil; ptr = ptr.next {
		if ptr.User.Name == username {
			return true, ptr
		}
	}
	return false, nil
}

// Utility Functions

// Hash function will take username as input and will be of type string --> int.
// Username will be used as search value while password will be checked for authentication.
// Implements Paul Larson's simple hash function for strings:
// https://stackoverflow.com/questions/98153/whats-the-best-hashing-algorithm-to-use-on-a-stl-string-when-using-hash-map/107657#107657

func hash(user string) int {
	hash := 0

	for key := range user {
		hash = hash*101 + int(user[key])
	}

	hash = hash % Hashbuckets

	return hash
}

// PrintSLLnoadmin function taken as argument for PrintHT(); prints username of non-admin users only (for non-admin users to set ticket assignee)
func PrintSLLnoadmin(SLL *UserNode) []string {
	result := []string{}
	if SLL == nil {
	} else {
		for ptr := SLL; ptr != nil; ptr = ptr.next {
			if !ptr.User.Admin {
				result = append(result, "Username: "+ptr.User.Name)
				result = append(result, "------------------------------")
			}
		}
	}
	return result
}

// PrintSLLusername function taken as argument for PrintHT(); prints usernames of admin and non-admin users (login screen ref)
func PrintSLLusername(SLL *UserNode) []string {
	result := []string{}
	if SLL == nil {
	} else {
		for ptr := SLL; ptr != nil; ptr = ptr.next {
			result = append(result, "Username: "+ptr.User.Name)
			if ptr.User.Admin {
				result = append(result, "Admin")
			} else {
				result = append(result, "Non-Admin")
			}
			result = append(result, "------------------------------")
		}
	}
	return result
}
