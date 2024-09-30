package data

/*validation_rules_start*/

import "github.com/jempe/whatsapp_backup/internal/validator"

//var (
//	ErrDuplicateContactTitleEn = errors.New("duplicate contact title (en)")
//	ErrDuplicateContactTitleEs = errors.New("duplicate contact title (es)")
//	ErrDuplicateContactTitleFr = errors.New("duplicate contact title (fr)")
//	ErrDuplicateContactURLEn   = errors.New("duplicate contact URL (en)")
//	ErrDuplicateContactURLEs   = errors.New("duplicate contact URL (es)")
//	ErrDuplicateContactURLFr   = errors.New("duplicate contact URL (fr)")
//	ErrDuplicateContactFolder  = errors.New("duplicate contact folder")
//)


func ValidateContact(v *validator.Validator, contact *Contact, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(contact.Name != "", "name", "must be provided")
	//v.Check(len(contact.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(contact.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func contactCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "contacts_title_en_key"`:
//		return ErrDuplicateContactTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "contacts_title_es_key"`:
//		return ErrDuplicateContactTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "contacts_title_fr_key"`:
//		return ErrDuplicateContactTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "contacts_url_en_key"`:
//		return ErrDuplicateContactURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "contacts_url_es_key"`:
//		return ErrDuplicateContactURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "contacts_url_fr_key"`:
//		return ErrDuplicateContactURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "contacts_channel_folder_key"`:
//		return ErrDuplicateContactFolder
	default:
		return err
	}
}
/*validation_rules_end*/

