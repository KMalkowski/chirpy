package database

import (
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Body     string `json:"body"`
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
}

type RefreshToken struct {
	Token     string    `json:"token"`
	UserId    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DBStructure struct {
	Chirps        map[int]Chirp  `json:"chirps"`
	Users         map[int]User   `json:"users"`
	RefreshTokens []RefreshToken `json:"refresh_tokens"`
}

func NewDB(path string) (*DB, error) {
	err := os.WriteFile(path, []byte{}, 0666)

	if err != nil {
		return &DB{}, nil
	}

	return &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}, nil
}

func (db *DB) ReadDatabase() (DBStructure, error) {
	file, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	content := DBStructure{}
	err = json.Unmarshal(file, &content)

	if err != nil {
		return DBStructure{}, fmt.Errorf("Could not decode the db")
	}

	return content, nil
}

func (db *DB) CreateChirp(body string, userId int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	file, err := os.ReadFile(db.path)
	if err != nil {
		return Chirp{}, err
	}

	content := DBStructure{}
	_ = json.Unmarshal(file, &content)

	chirpId := len(content.Chirps) + 1
	newChirp := Chirp{Body: body, Id: chirpId, AuthorId: userId}

	if len(content.Chirps) > 0 {
		content.Chirps[len(content.Chirps)] = newChirp
	} else {
		content.Chirps = make(map[int]Chirp)
		content.Chirps[len(content.Chirps)] = newChirp
	}

	encodedContent, err := json.Marshal(content)
	if err != nil {
		return Chirp{}, err
	}

	os.WriteFile(db.path, encodedContent, 0666)
	return content.Chirps[len(content.Chirps)-1], nil
}

func (db *DB) DeleteChirp(id int, userId int) error {
	db.mux.RLock()
	defer db.mux.RUnlock()

	database, err := db.ReadDatabase()
	if err != nil {
		return err
	}

	theChirp := Chirp{}
	index := 0
	for i, c := range database.Chirps {
		if c.Id == id {
			theChirp = database.Chirps[i]
			index = i
			break
		}
	}

	if (Chirp{}) == theChirp {
		return fmt.Errorf("chirp not found")
	}

	if theChirp.AuthorId != userId {
		return fmt.Errorf("user in not the author")
	}

	delete(database.Chirps, index)

	newDb, err := json.Marshal(database)
	if err != nil {
		return err
	}
	os.WriteFile(db.path, newDb, 0666)

	return nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	file, err := os.ReadFile(db.path)
	if err != nil {
		return []Chirp{}, err
	}

	content := DBStructure{}
	err = json.Unmarshal(file, &content)

	if err != nil {
		return []Chirp{}, err
	}

	v := make([]Chirp, 0, len(content.Chirps))

	for _, value := range content.Chirps {
		v = append(v, value)
	}

	return v, nil
}

type User struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mux.RLock()
	db.mux.RUnlock()

	_, err := mail.ParseAddress(email)
	if err != nil {
		return User{}, fmt.Errorf("Email in invalid")
	}

	if len(password) < 4 {
		return User{}, fmt.Errorf("Password has to be at least 4 characters")
	}

	dbfile, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatalln("err")
		return User{}, err
	}

	dbContent := DBStructure{}
	_ = json.Unmarshal(dbfile, &dbContent)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		return User{}, err
	}

	newUser := User{Id: len(dbContent.Users) + 1, Email: email, Password: string(hashedPassword), IsChirpyRed: false}
	if len(dbContent.Users) < 1 {
		dbContent.Users = make(map[int]User)
	}
	dbContent.Users[len(dbContent.Users)] = newUser

	newDb, err := json.Marshal(dbContent)
	if err != nil {
		return User{}, err
	}
	os.WriteFile(db.path, newDb, 0666)

	return newUser, nil
}

func (db *DB) UpdateUser(id int, email string, password string) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	database, err := db.ReadDatabase()
	if err != nil {
		return User{}, err
	}

	var index int
	newUser := User{}
	for i, user := range database.Users {
		if user.Id == id {
			newUser = user
			index = i
			break
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		return User{}, err
	}

	newUser.Email = email
	newUser.Password = string(hashedPassword)

	database.Users[index] = newUser
	newDb, err := json.Marshal(database)
	if err != nil {
		return User{}, err
	}
	os.WriteFile(db.path, newDb, 0666)

	return newUser, nil
}

func (db *DB) UpgradeUser(id int, isChirpyRed bool) error {
	db.mux.RLock()
	defer db.mux.RUnlock()

	database, err := db.ReadDatabase()
	if err != nil {
		return err
	}

	var index int
	newUser := User{}
	for i, user := range database.Users {
		if user.Id == id {
			newUser = user
			index = i
			break
		}
	}

	if (User{}) == newUser {
		log.Println("user not found in db to upgrade")
		return fmt.Errorf("user not found")
	}

	newUser.IsChirpyRed = isChirpyRed

	database.Users[index] = newUser
	newDb, err := json.Marshal(database)
	if err != nil {
		return err
	}
	os.WriteFile(db.path, newDb, 0666)

	return nil
}

func (db *DB) AddRefreshToken(token string, expiresAt time.Time, userId int) (RefreshToken, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	database, err := db.ReadDatabase()
	if err != nil {
		return RefreshToken{}, err
	}

	newToken := RefreshToken{Token: token, ExpiresAt: expiresAt, UserId: userId}
	newTokensSlice := []RefreshToken{}
	if len(database.RefreshTokens) < 1 {
		database.RefreshTokens = []RefreshToken{}
		newTokensSlice = append(database.RefreshTokens, newToken)
	} else {
		newTokensSlice = append(database.RefreshTokens, newToken)
	}

	database.RefreshTokens = newTokensSlice

	newDb, err := json.Marshal(database)
	if err != nil {
		return RefreshToken{}, err
	}
	os.WriteFile(db.path, newDb, 0666)

	return newToken, nil
}

func RemoveIndex(s []RefreshToken, index int) []RefreshToken {
	return append(s[:index], s[index+1:]...)
}

func (db *DB) RevokeRefreshToken(authToken string) error {
	database, err := db.ReadDatabase()
	if err != nil {
		return err
	}

	for i, token := range database.RefreshTokens {
		if token.Token == authToken {
			database.RefreshTokens = RemoveIndex(database.RefreshTokens, i)
			break
		}
	}

	newDb, err := json.Marshal(database)
	if err != nil {
		return err
	}
	os.WriteFile(db.path, newDb, 0666)

	return nil
}
