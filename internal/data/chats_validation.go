package data

/*validation_rules_start*/

import "github.com/jempe/whatsapp_backup/internal/validator"

//var (
//	ErrDuplicateChatTitleEn = errors.New("duplicate chat title (en)")
//	ErrDuplicateChatTitleEs = errors.New("duplicate chat title (es)")
//	ErrDuplicateChatTitleFr = errors.New("duplicate chat title (fr)")
//	ErrDuplicateChatURLEn   = errors.New("duplicate chat URL (en)")
//	ErrDuplicateChatURLEs   = errors.New("duplicate chat URL (es)")
//	ErrDuplicateChatURLFr   = errors.New("duplicate chat URL (fr)")
//	ErrDuplicateChatFolder  = errors.New("duplicate chat folder")
//)


func ValidateChat(v *validator.Validator, chat *Chat, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(chat.Name != "", "name", "must be provided")
	//v.Check(len(chat.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(chat.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func chatCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "chats_title_en_key"`:
//		return ErrDuplicateChatTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "chats_title_es_key"`:
//		return ErrDuplicateChatTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "chats_title_fr_key"`:
//		return ErrDuplicateChatTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "chats_url_en_key"`:
//		return ErrDuplicateChatURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "chats_url_es_key"`:
//		return ErrDuplicateChatURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "chats_url_fr_key"`:
//		return ErrDuplicateChatURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "chats_channel_folder_key"`:
//		return ErrDuplicateChatFolder
	default:
		return err
	}
}
/*validation_rules_end*/

