package data

/*validation_rules_start*/

import "github.com/jempe/whatsapp_backup/internal/validator"

//var (
//	ErrDuplicatePhraseTitleEn = errors.New("duplicate phrase title (en)")
//	ErrDuplicatePhraseTitleEs = errors.New("duplicate phrase title (es)")
//	ErrDuplicatePhraseTitleFr = errors.New("duplicate phrase title (fr)")
//	ErrDuplicatePhraseURLEn   = errors.New("duplicate phrase URL (en)")
//	ErrDuplicatePhraseURLEs   = errors.New("duplicate phrase URL (es)")
//	ErrDuplicatePhraseURLFr   = errors.New("duplicate phrase URL (fr)")
//	ErrDuplicatePhraseFolder  = errors.New("duplicate phrase folder")
//)


func ValidatePhrase(v *validator.Validator, phrase *Phrase, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(phrase.Name != "", "name", "must be provided")
	//v.Check(len(phrase.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(phrase.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func phraseCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "phrases_title_en_key"`:
//		return ErrDuplicatePhraseTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "phrases_title_es_key"`:
//		return ErrDuplicatePhraseTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "phrases_title_fr_key"`:
//		return ErrDuplicatePhraseTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "phrases_url_en_key"`:
//		return ErrDuplicatePhraseURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "phrases_url_es_key"`:
//		return ErrDuplicatePhraseURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "phrases_url_fr_key"`:
//		return ErrDuplicatePhraseURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "phrases_channel_folder_key"`:
//		return ErrDuplicatePhraseFolder
	default:
		return err
	}
}
/*validation_rules_end*/

