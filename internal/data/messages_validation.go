package data

/*validation_rules_start*/

import "github.com/jempe/whatsapp_backup/internal/validator"

//var (
//	ErrDuplicateMessageTitleEn = errors.New("duplicate message title (en)")
//	ErrDuplicateMessageTitleEs = errors.New("duplicate message title (es)")
//	ErrDuplicateMessageTitleFr = errors.New("duplicate message title (fr)")
//	ErrDuplicateMessageURLEn   = errors.New("duplicate message URL (en)")
//	ErrDuplicateMessageURLEs   = errors.New("duplicate message URL (es)")
//	ErrDuplicateMessageURLFr   = errors.New("duplicate message URL (fr)")
//	ErrDuplicateMessageFolder  = errors.New("duplicate message folder")
//)


func ValidateMessage(v *validator.Validator, message *Message, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(message.Name != "", "name", "must be provided")
	//v.Check(len(message.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(message.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func messageCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "messages_title_en_key"`:
//		return ErrDuplicateMessageTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "messages_title_es_key"`:
//		return ErrDuplicateMessageTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "messages_title_fr_key"`:
//		return ErrDuplicateMessageTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "messages_url_en_key"`:
//		return ErrDuplicateMessageURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "messages_url_es_key"`:
//		return ErrDuplicateMessageURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "messages_url_fr_key"`:
//		return ErrDuplicateMessageURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "messages_channel_folder_key"`:
//		return ErrDuplicateMessageFolder
	default:
		return err
	}
}
/*validation_rules_end*/

