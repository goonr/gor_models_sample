// Package models includes the functions on the model Picture.
package models

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

// set flags to output more detailed log
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Picture struct {
	Id            int64     `json:"id,omitempty" db:"id" valid:"-"`
	Name          string    `json:"name,omitempty" db:"name" valid:"-"`
	Url           string    `json:"url,omitempty" db:"url" valid:"-"`
	ImageableId   int64     `json:"imageable_id,omitempty" db:"imageable_id" valid:"-"`
	ImageableType string    `json:"imageable_type,omitempty" db:"imageable_type" valid:"-"`
	CreatedAt     time.Time `json:"created_at,omitempty" db:"created_at" valid:"-"`
	UpdatedAt     time.Time `json:"updated_at,omitempty" db:"updated_at" valid:"-"`
}

// DataStruct for the pagination
type PicturePage struct {
	WhereString string
	WhereParams []interface{}
	Order       map[string]string
	FirstId     int64
	LastId      int64
	PageNum     int
	PerPage     int
	TotalPages  int
	TotalItems  int64
	orderStr    string
}

// Current get the current page of PicturePage object for pagination.
func (_p *PicturePage) Current() ([]Picture, error) {
	if _, exist := _p.Order["id"]; !exist {
		return nil, errors.New("No id order specified in Order map")
	}
	err := _p.buildPageCount()
	if err != nil {
		return nil, fmt.Errorf("Calculate page count error: %v", err)
	}
	if _p.orderStr == "" {
		_p.buildOrder()
	}
	idStr, idParams := _p.buildIdRestrict("current")
	whereStr := fmt.Sprintf("%s %s %s LIMIT %v", _p.WhereString, idStr, _p.orderStr, _p.PerPage)
	whereParams := []interface{}{}
	whereParams = append(append(whereParams, _p.WhereParams...), idParams...)
	pictures, err := FindPicturesWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(pictures) != 0 {
		_p.FirstId, _p.LastId = pictures[0].Id, pictures[len(pictures)-1].Id
	}
	return pictures, nil
}

// Previous get the previous page of PicturePage object for pagination.
func (_p *PicturePage) Previous() ([]Picture, error) {
	if _p.PageNum == 0 {
		return nil, errors.New("This's the first page, no previous page yet")
	}
	if _, exist := _p.Order["id"]; !exist {
		return nil, errors.New("No id order specified in Order map")
	}
	err := _p.buildPageCount()
	if err != nil {
		return nil, fmt.Errorf("Calculate page count error: %v", err)
	}
	if _p.orderStr == "" {
		_p.buildOrder()
	}
	idStr, idParams := _p.buildIdRestrict("previous")
	whereStr := fmt.Sprintf("%s %s %s LIMIT %v", _p.WhereString, idStr, _p.orderStr, _p.PerPage)
	whereParams := []interface{}{}
	whereParams = append(append(whereParams, _p.WhereParams...), idParams...)
	pictures, err := FindPicturesWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(pictures) != 0 {
		_p.FirstId, _p.LastId = pictures[0].Id, pictures[len(pictures)-1].Id
	}
	_p.PageNum -= 1
	return pictures, nil
}

// Next get the next page of PicturePage object for pagination.
func (_p *PicturePage) Next() ([]Picture, error) {
	if _p.PageNum == _p.TotalPages-1 {
		return nil, errors.New("This's the last page, no next page yet")
	}
	if _, exist := _p.Order["id"]; !exist {
		return nil, errors.New("No id order specified in Order map")
	}
	err := _p.buildPageCount()
	if err != nil {
		return nil, fmt.Errorf("Calculate page count error: %v", err)
	}
	if _p.orderStr == "" {
		_p.buildOrder()
	}
	idStr, idParams := _p.buildIdRestrict("next")
	whereStr := fmt.Sprintf("%s %s %s LIMIT %v", _p.WhereString, idStr, _p.orderStr, _p.PerPage)
	whereParams := []interface{}{}
	whereParams = append(append(whereParams, _p.WhereParams...), idParams...)
	pictures, err := FindPicturesWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(pictures) != 0 {
		_p.FirstId, _p.LastId = pictures[0].Id, pictures[len(pictures)-1].Id
	}
	_p.PageNum += 1
	return pictures, nil
}

// GetPage is a helper function for the PicturePage object to return a corresponding page due to
// the parameter passed in, i.e. one of "previous, current or next".
func (_p *PicturePage) GetPage(direction string) (ps []Picture, err error) {
	switch direction {
	case "previous":
		ps, _ = _p.Previous()
	case "next":
		ps, _ = _p.Next()
	case "current":
		ps, _ = _p.Current()
	default:
		return nil, errors.New("Error: wrong dircetion! None of previous, current or next!")
	}
	return
}

// buildOrder is for PicturePage object to build a SQL ORDER BY clause.
func (_p *PicturePage) buildOrder() {
	tempList := []string{}
	for k, v := range _p.Order {
		tempList = append(tempList, fmt.Sprintf("%v %v", k, v))
	}
	_p.orderStr = " ORDER BY " + strings.Join(tempList, ", ")
}

// buildIdRestrict is for PicturePage object to build a SQL clause for ID restriction,
// implementing a simple keyset style pagination.
func (_p *PicturePage) buildIdRestrict(direction string) (idStr string, idParams []interface{}) {
	switch direction {
	case "previous":
		if strings.ToLower(_p.Order["id"]) == "desc" {
			idStr += "id > ? "
			idParams = append(idParams, _p.FirstId)
		} else {
			idStr += "id < ? "
			idParams = append(idParams, _p.FirstId)
		}
	case "current":
		// trick to make Where function work
		if _p.PageNum == 0 && _p.FirstId == 0 && _p.LastId == 0 {
			idStr += "id > ? "
			idParams = append(idParams, 0)
		} else {
			if strings.ToLower(_p.Order["id"]) == "desc" {
				idStr += "id <= ? AND id >= ? "
				idParams = append(idParams, _p.FirstId, _p.LastId)
			} else {
				idStr += "id >= ? AND id <= ? "
				idParams = append(idParams, _p.FirstId, _p.LastId)
			}
		}
	case "next":
		if strings.ToLower(_p.Order["id"]) == "desc" {
			idStr += "id < ? "
			idParams = append(idParams, _p.LastId)
		} else {
			idStr += "id > ? "
			idParams = append(idParams, _p.LastId)
		}
	}
	if _p.WhereString != "" {
		idStr = " AND " + idStr
	}
	return
}

// buildPageCount calculate the TotalItems/TotalPages for the PicturePage object.
func (_p *PicturePage) buildPageCount() error {
	count, err := PictureCountWhere(_p.WhereString, _p.WhereParams...)
	if err != nil {
		return err
	}
	_p.TotalItems = count
	if _p.PerPage == 0 {
		_p.PerPage = 10
	}
	_p.TotalPages = int(math.Ceil(float64(_p.TotalItems) / float64(_p.PerPage)))
	return nil
}

// FindPicture find a single picture by an ID.
func FindPicture(id int64) (*Picture, error) {
	if id == 0 {
		return nil, errors.New("Invalid ID: it can't be zero")
	}
	_picture := Picture{}
	err := DB.Get(&_picture, DB.Rebind(`SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures WHERE pictures.id = ? LIMIT 1`), id)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_picture, nil
}

// FirstPicture find the first one picture by ID ASC order.
func FirstPicture() (*Picture, error) {
	_picture := Picture{}
	err := DB.Get(&_picture, DB.Rebind(`SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures ORDER BY pictures.id ASC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_picture, nil
}

// FirstPictures find the first N pictures by ID ASC order.
func FirstPictures(n uint32) ([]Picture, error) {
	_pictures := []Picture{}
	sql := fmt.Sprintf("SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures ORDER BY pictures.id ASC LIMIT %v", n)
	err := DB.Select(&_pictures, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _pictures, nil
}

// LastPicture find the last one picture by ID DESC order.
func LastPicture() (*Picture, error) {
	_picture := Picture{}
	err := DB.Get(&_picture, DB.Rebind(`SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures ORDER BY pictures.id DESC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_picture, nil
}

// LastPictures find the last N pictures by ID DESC order.
func LastPictures(n uint32) ([]Picture, error) {
	_pictures := []Picture{}
	sql := fmt.Sprintf("SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures ORDER BY pictures.id DESC LIMIT %v", n)
	err := DB.Select(&_pictures, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _pictures, nil
}

// FindPictures find one or more pictures by the given ID(s).
func FindPictures(ids ...int64) ([]Picture, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return nil, errors.New(msg)
	}
	_pictures := []Picture{}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := DB.Rebind(fmt.Sprintf(`SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures WHERE pictures.id IN (?%s)`, idsHolder))
	idsT := []interface{}{}
	for _, id := range ids {
		idsT = append(idsT, interface{}(id))
	}
	err := DB.Select(&_pictures, sql, idsT...)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _pictures, nil
}

// FindPictureBy find a single picture by a field name and a value.
func FindPictureBy(field string, val interface{}) (*Picture, error) {
	_picture := Picture{}
	sqlFmt := `SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures WHERE %s = ? LIMIT 1`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err := DB.Get(&_picture, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_picture, nil
}

// FindPicturesBy find all pictures by a field name and a value.
func FindPicturesBy(field string, val interface{}) (_pictures []Picture, err error) {
	sqlFmt := `SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures WHERE %s = ?`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err = DB.Select(&_pictures, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _pictures, nil
}

// AllPictures get all the Picture records.
func AllPictures() (pictures []Picture, err error) {
	err = DB.Select(&pictures, "SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return pictures, nil
}

// PictureCount get the count of all the Picture records.
func PictureCount() (c int64, err error) {
	err = DB.Get(&c, "SELECT count(*) FROM pictures")
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return c, nil
}

// PictureCountWhere get the count of all the Picture records with a where clause.
func PictureCountWhere(where string, args ...interface{}) (c int64, err error) {
	sql := "SELECT count(*) FROM pictures"
	if len(where) > 0 {
		sql = sql + " WHERE " + where
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return 0, err
	}
	err = stmt.Get(&c, args...)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return c, nil
}

// PictureIncludesWhere get the Picture associated models records, currently it's not same as the corresponding "includes" function but "preload" instead in Ruby on Rails. It means that the "sql" should be restricted on Picture model.
func PictureIncludesWhere(assocs []string, sql string, args ...interface{}) (_pictures []Picture, err error) {
	_pictures, err = FindPicturesWhere(sql, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(assocs) == 0 {
		log.Println("No associated fields ard specified")
		return _pictures, err
	}
	if len(_pictures) <= 0 {
		return nil, errors.New("No results available")
	}
	ids := make([]interface{}, len(_pictures))
	for _, v := range _pictures {
		ids = append(ids, interface{}(v.Id))
	}
	return _pictures, nil
}

// PictureIds get all the IDs of Picture records.
func PictureIds() (ids []int64, err error) {
	err = DB.Select(&ids, "SELECT id FROM pictures")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return ids, nil
}

// PictureIdsWhere get all the IDs of Picture records by where restriction.
func PictureIdsWhere(where string, args ...interface{}) ([]int64, error) {
	ids, err := PictureIntCol("id", where, args...)
	return ids, err
}

// PictureIntCol get some int64 typed column of Picture by where restriction.
func PictureIntCol(col, where string, args ...interface{}) (intColRecs []int64, err error) {
	sql := "SELECT " + col + " FROM pictures"
	if len(where) > 0 {
		sql = sql + " WHERE " + where
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&intColRecs, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return intColRecs, nil
}

// PictureStrCol get some string typed column of Picture by where restriction.
func PictureStrCol(col, where string, args ...interface{}) (strColRecs []string, err error) {
	sql := "SELECT " + col + " FROM pictures"
	if len(where) > 0 {
		sql = sql + " WHERE " + where
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&strColRecs, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return strColRecs, nil
}

// FindPicturesWhere query use a partial SQL clause that usually following after WHERE
// with placeholders, eg: FindUsersWhere("first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindPicturesWhere(where string, args ...interface{}) (pictures []Picture, err error) {
	sql := "SELECT COALESCE(pictures.name, '') AS name, COALESCE(pictures.url, '') AS url, COALESCE(pictures.imageable_id, 0) AS imageable_id, COALESCE(pictures.imageable_type, '') AS imageable_type, pictures.id, pictures.created_at, pictures.updated_at FROM pictures"
	if len(where) > 0 {
		sql = sql + " WHERE " + where
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&pictures, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return pictures, nil
}

// FindPictureBySql query use a complete SQL clause
// with placeholders, eg: FindUserBySql("SELECT * FROM users WHERE first_name = ? AND age > ? ORDER BY DESC LIMIT 1", "John", 18)
// will return only One record in the table "users" whose first_name is "John" and age elder than 18.
func FindPictureBySql(sql string, args ...interface{}) (*Picture, error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_picture := &Picture{}
	err = stmt.Get(_picture, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return _picture, nil
}

// FindPicturesBySql query use a complete SQL clause
// with placeholders, eg: FindUsersBySql("SELECT * FROM users WHERE first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindPicturesBySql(sql string, args ...interface{}) (pictures []Picture, err error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&pictures, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return pictures, nil
}

// CreatePicture use a named params to create a single Picture record.
// A named params is key-value map like map[string]interface{}{"first_name": "John", "age": 23} .
func CreatePicture(am map[string]interface{}) (int64, error) {
	if len(am) == 0 {
		return 0, fmt.Errorf("Zero key in the attributes map!")
	}
	t := time.Now()
	for _, v := range []string{"created_at", "updated_at"} {
		if am[v] == nil {
			am[v] = t
		}
	}
	keys := allKeys(am)
	sqlFmt := `INSERT INTO pictures (%s) VALUES (%s)`
	sql := fmt.Sprintf(sqlFmt, strings.Join(keys, ","), ":"+strings.Join(keys, ",:"))
	result, err := DB.NamedExec(sql, am)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return lastId, nil
}

// Create is a method for Picture to create a record.
func (_picture *Picture) Create() (int64, error) {
	ok, err := govalidator.ValidateStruct(_picture)
	if !ok {
		errMsg := "Validate Picture struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Picture struct error: " + err.Error()
		}
		log.Println(errMsg)
		return 0, errors.New(errMsg)
	}
	t := time.Now()
	_picture.CreatedAt = t
	_picture.UpdatedAt = t
	sql := `INSERT INTO pictures (name,url,imageable_id,imageable_type,created_at,updated_at) VALUES (:name,:url,:imageable_id,:imageable_type,:created_at,:updated_at)`
	result, err := DB.NamedExec(sql, _picture)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return lastId, nil
}

// Destroy is method used for a Picture object to be destroyed.
func (_picture *Picture) Destroy() error {
	if _picture.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := DestroyPicture(_picture.Id)
	return err
}

// DestroyPicture will destroy a Picture record specified by the id parameter.
func DestroyPicture(id int64) error {
	stmt, err := DB.Preparex(DB.Rebind(`DELETE FROM pictures WHERE id = ?`))
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

// DestroyPictures will destroy Picture records those specified by the ids parameters.
func DestroyPictures(ids ...int64) (int64, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return 0, errors.New(msg)
	}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := fmt.Sprintf(`DELETE FROM pictures WHERE id IN (?%s)`, idsHolder)
	idsT := []interface{}{}
	for _, id := range ids {
		idsT = append(idsT, interface{}(id))
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	result, err := stmt.Exec(idsT...)
	if err != nil {
		return 0, err
	}
	cnt, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

// DestroyPicturesWhere delete records by a where clause restriction.
// e.g. DestroyPicturesWhere("name = ?", "John")
// And this func will not call the association dependent action
func DestroyPicturesWhere(where string, args ...interface{}) (int64, error) {
	sql := `DELETE FROM pictures WHERE `
	if len(where) > 0 {
		sql = sql + where
	} else {
		return 0, errors.New("No WHERE conditions provided")
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}
	cnt, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

// Save method is used for a Picture object to update an existed record mainly.
// If no id provided a new record will be created. FIXME: A UPSERT action will be implemented further.
func (_picture *Picture) Save() error {
	ok, err := govalidator.ValidateStruct(_picture)
	if !ok {
		errMsg := "Validate Picture struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Picture struct error: " + err.Error()
		}
		log.Println(errMsg)
		return errors.New(errMsg)
	}
	if _picture.Id == 0 {
		_, err = _picture.Create()
		return err
	}
	_picture.UpdatedAt = time.Now()
	sqlFmt := `UPDATE pictures SET %s WHERE id = %v`
	sqlStr := fmt.Sprintf(sqlFmt, "name = :name, url = :url, imageable_id = :imageable_id, imageable_type = :imageable_type, updated_at = :updated_at", _picture.Id)
	_, err = DB.NamedExec(sqlStr, _picture)
	return err
}

// UpdatePicture is used to update a record with a id and map[string]interface{} typed key-value parameters.
func UpdatePicture(id int64, am map[string]interface{}) error {
	if len(am) == 0 {
		return errors.New("Zero key in the attributes map!")
	}
	am["updated_at"] = time.Now()
	keys := allKeys(am)
	sqlFmt := `UPDATE pictures SET %s WHERE id = %v`
	setKeysArr := []string{}
	for _, v := range keys {
		s := fmt.Sprintf(" %s = :%s", v, v)
		setKeysArr = append(setKeysArr, s)
	}
	sqlStr := fmt.Sprintf(sqlFmt, strings.Join(setKeysArr, ", "), id)
	_, err := DB.NamedExec(sqlStr, am)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Update is a method used to update a Picture record with the map[string]interface{} typed key-value parameters.
func (_picture *Picture) Update(am map[string]interface{}) error {
	if _picture.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePicture(_picture.Id, am)
	return err
}

// UpdateAttributes method is supposed to be used to update Picture records as corresponding update_attributes in Ruby on Rails.
func (_picture *Picture) UpdateAttributes(am map[string]interface{}) error {
	if _picture.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePicture(_picture.Id, am)
	return err
}

// UpdateColumns method is supposed to be used to update Picture records as corresponding update_columns in Ruby on Rails.
func (_picture *Picture) UpdateColumns(am map[string]interface{}) error {
	if _picture.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePicture(_picture.Id, am)
	return err
}

// UpdatePicturesBySql is used to update Picture records by a SQL clause
// using the '?' binding syntax.
func UpdatePicturesBySql(sql string, args ...interface{}) (int64, error) {
	if sql == "" {
		return 0, errors.New("A blank SQL clause")
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}
	cnt, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return cnt, nil
}
