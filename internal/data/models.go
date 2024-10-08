package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Contacts ContactModel
	Messages MessageModel
	Chats    ChatModel
	Phrases  PhraseModel
	Users    UserModel
	Tokens   TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Contacts: ContactModel{DB: db},
		Messages: MessageModel{DB: db},
		Chats:    ChatModel{DB: db},
		Phrases:  PhraseModel{DB: db},
		Users:    UserModel{DB: db},
		Tokens:   TokenModel{DB: db},
	}
}
