// Package models includes the functions on the model Physician.
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

type Physician struct {
	Id           int64         `json:"id,omitempty" db:"id" valid:"-"`
	Name         string        `json:"name,omitempty" db:"name" valid:"required,length(6|15)"`
	CreatedAt    time.Time     `json:"created_at,omitempty" db:"created_at" valid:"-"`
	UpdatedAt    time.Time     `json:"updated_at,omitempty" db:"updated_at" valid:"-"`
	Introduction string        `json:"introduction,omitempty" db:"introduction" valid:"required"`
	Appointments []Appointment `json:"appointments,omitempty" db:"appointments" valid:"-"`
	Patients     []Patient     `json:"patients,omitempty" db:"patients" valid:"-"`
	Pictures     []Picture     `json:"pictures,omitempty" db:"pictures" valid:"-"`
}

// DataStruct for the pagination
type PhysicianPage struct {
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

// Current get the current page of PhysicianPage object for pagination.
func (_p *PhysicianPage) Current() ([]Physician, error) {
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
	physicians, err := FindPhysiciansWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(physicians) != 0 {
		_p.FirstId, _p.LastId = physicians[0].Id, physicians[len(physicians)-1].Id
	}
	return physicians, nil
}

// Previous get the previous page of PhysicianPage object for pagination.
func (_p *PhysicianPage) Previous() ([]Physician, error) {
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
	physicians, err := FindPhysiciansWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(physicians) != 0 {
		_p.FirstId, _p.LastId = physicians[0].Id, physicians[len(physicians)-1].Id
	}
	_p.PageNum -= 1
	return physicians, nil
}

// Next get the next page of PhysicianPage object for pagination.
func (_p *PhysicianPage) Next() ([]Physician, error) {
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
	physicians, err := FindPhysiciansWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(physicians) != 0 {
		_p.FirstId, _p.LastId = physicians[0].Id, physicians[len(physicians)-1].Id
	}
	_p.PageNum += 1
	return physicians, nil
}

// GetPage is a helper function for the PhysicianPage object to return a corresponding page due to
// the parameter passed in, i.e. one of "previous, current or next".
func (_p *PhysicianPage) GetPage(direction string) (ps []Physician, err error) {
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

// buildOrder is for PhysicianPage object to build a SQL ORDER BY clause.
func (_p *PhysicianPage) buildOrder() {
	tempList := []string{}
	for k, v := range _p.Order {
		tempList = append(tempList, fmt.Sprintf("%v %v", k, v))
	}
	_p.orderStr = " ORDER BY " + strings.Join(tempList, ", ")
}

// buildIdRestrict is for PhysicianPage object to build a SQL clause for ID restriction,
// implementing a simple keyset style pagination.
func (_p *PhysicianPage) buildIdRestrict(direction string) (idStr string, idParams []interface{}) {
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

// buildPageCount calculate the TotalItems/TotalPages for the PhysicianPage object.
func (_p *PhysicianPage) buildPageCount() error {
	count, err := PhysicianCountWhere(_p.WhereString, _p.WhereParams...)
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

// FindPhysician find a single physician by an ID.
func FindPhysician(id int64) (*Physician, error) {
	if id == 0 {
		return nil, errors.New("Invalid ID: it can't be zero")
	}
	_physician := Physician{}
	err := DB.Get(&_physician, DB.Rebind(`SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians WHERE physicians.id = ? LIMIT 1`), id)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_physician, nil
}

// FirstPhysician find the first one physician by ID ASC order.
func FirstPhysician() (*Physician, error) {
	_physician := Physician{}
	err := DB.Get(&_physician, DB.Rebind(`SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians ORDER BY physicians.id ASC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_physician, nil
}

// FirstPhysicians find the first N physicians by ID ASC order.
func FirstPhysicians(n uint32) ([]Physician, error) {
	_physicians := []Physician{}
	sql := fmt.Sprintf("SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians ORDER BY physicians.id ASC LIMIT %v", n)
	err := DB.Select(&_physicians, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _physicians, nil
}

// LastPhysician find the last one physician by ID DESC order.
func LastPhysician() (*Physician, error) {
	_physician := Physician{}
	err := DB.Get(&_physician, DB.Rebind(`SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians ORDER BY physicians.id DESC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_physician, nil
}

// LastPhysicians find the last N physicians by ID DESC order.
func LastPhysicians(n uint32) ([]Physician, error) {
	_physicians := []Physician{}
	sql := fmt.Sprintf("SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians ORDER BY physicians.id DESC LIMIT %v", n)
	err := DB.Select(&_physicians, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _physicians, nil
}

// FindPhysicians find one or more physicians by the given ID(s).
func FindPhysicians(ids ...int64) ([]Physician, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return nil, errors.New(msg)
	}
	_physicians := []Physician{}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := DB.Rebind(fmt.Sprintf(`SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians WHERE physicians.id IN (?%s)`, idsHolder))
	idsT := []interface{}{}
	for _, id := range ids {
		idsT = append(idsT, interface{}(id))
	}
	err := DB.Select(&_physicians, sql, idsT...)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _physicians, nil
}

// FindPhysicianBy find a single physician by a field name and a value.
func FindPhysicianBy(field string, val interface{}) (*Physician, error) {
	_physician := Physician{}
	sqlFmt := `SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians WHERE %s = ? LIMIT 1`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err := DB.Get(&_physician, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_physician, nil
}

// FindPhysiciansBy find all physicians by a field name and a value.
func FindPhysiciansBy(field string, val interface{}) (_physicians []Physician, err error) {
	sqlFmt := `SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians WHERE %s = ?`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err = DB.Select(&_physicians, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _physicians, nil
}

// AllPhysicians get all the Physician records.
func AllPhysicians() (physicians []Physician, err error) {
	err = DB.Select(&physicians, "SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return physicians, nil
}

// PhysicianCount get the count of all the Physician records.
func PhysicianCount() (c int64, err error) {
	err = DB.Get(&c, "SELECT count(*) FROM physicians")
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return c, nil
}

// PhysicianCountWhere get the count of all the Physician records with a where clause.
func PhysicianCountWhere(where string, args ...interface{}) (c int64, err error) {
	sql := "SELECT count(*) FROM physicians"
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

// PhysicianIncludesWhere get the Physician associated models records, currently it's not same as the corresponding "includes" function but "preload" instead in Ruby on Rails. It means that the "sql" should be restricted on Physician model.
func PhysicianIncludesWhere(assocs []string, sql string, args ...interface{}) (_physicians []Physician, err error) {
	_physicians, err = FindPhysiciansWhere(sql, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(assocs) == 0 {
		log.Println("No associated fields ard specified")
		return _physicians, err
	}
	if len(_physicians) <= 0 {
		return nil, errors.New("No results available")
	}
	ids := make([]interface{}, len(_physicians))
	for _, v := range _physicians {
		ids = append(ids, interface{}(v.Id))
	}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	for _, assoc := range assocs {
		switch assoc {
		case "appointments":
			where := fmt.Sprintf("physician_id IN (?%s)", idsHolder)
			_appointments, err := FindAppointmentsWhere(where, ids...)
			if err != nil {
				log.Printf("Error when query associated objects: %v\n", assoc)
				continue
			}
			for _, vv := range _appointments {
				for i, vvv := range _physicians {
					if vv.PhysicianId == vvv.Id {
						vvv.Appointments = append(vvv.Appointments, vv)
					}
					_physicians[i].Appointments = vvv.Appointments
				}
			}
		case "patients":
			// FIXME: optimize the query
			for i, vvv := range _physicians {
				_patients, err := PhysicianGetPatients(vvv.Id)
				if err != nil {
					continue
				}
				vvv.Patients = _patients
				_physicians[i] = vvv
			}
		case "pictures":
			// FIXME: optimize the query
			for i, vvv := range _physicians {
				_pictures, err := PhysicianGetPictures(vvv.Id)
				if err != nil {
					continue
				}
				vvv.Pictures = _pictures
				_physicians[i] = vvv
			}
		}
	}
	return _physicians, nil
}

// PhysicianIds get all the IDs of Physician records.
func PhysicianIds() (ids []int64, err error) {
	err = DB.Select(&ids, "SELECT id FROM physicians")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return ids, nil
}

// PhysicianIdsWhere get all the IDs of Physician records by where restriction.
func PhysicianIdsWhere(where string, args ...interface{}) ([]int64, error) {
	ids, err := PhysicianIntCol("id", where, args...)
	return ids, err
}

// PhysicianIntCol get some int64 typed column of Physician by where restriction.
func PhysicianIntCol(col, where string, args ...interface{}) (intColRecs []int64, err error) {
	sql := "SELECT " + col + " FROM physicians"
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

// PhysicianStrCol get some string typed column of Physician by where restriction.
func PhysicianStrCol(col, where string, args ...interface{}) (strColRecs []string, err error) {
	sql := "SELECT " + col + " FROM physicians"
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

// FindPhysiciansWhere query use a partial SQL clause that usually following after WHERE
// with placeholders, eg: FindUsersWhere("first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindPhysiciansWhere(where string, args ...interface{}) (physicians []Physician, err error) {
	sql := "SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at FROM physicians"
	if len(where) > 0 {
		sql = sql + " WHERE " + where
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&physicians, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return physicians, nil
}

// FindPhysicianBySql query use a complete SQL clause
// with placeholders, eg: FindUserBySql("SELECT * FROM users WHERE first_name = ? AND age > ? ORDER BY DESC LIMIT 1", "John", 18)
// will return only One record in the table "users" whose first_name is "John" and age elder than 18.
func FindPhysicianBySql(sql string, args ...interface{}) (*Physician, error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_physician := &Physician{}
	err = stmt.Get(_physician, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return _physician, nil
}

// FindPhysiciansBySql query use a complete SQL clause
// with placeholders, eg: FindUsersBySql("SELECT * FROM users WHERE first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindPhysiciansBySql(sql string, args ...interface{}) (physicians []Physician, err error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&physicians, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return physicians, nil
}

// CreatePhysician use a named params to create a single Physician record.
// A named params is key-value map like map[string]interface{}{"first_name": "John", "age": 23} .
func CreatePhysician(am map[string]interface{}) (int64, error) {
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
	sqlFmt := `INSERT INTO physicians (%s) VALUES (%s)`
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

// Create is a method for Physician to create a record.
func (_physician *Physician) Create() (int64, error) {
	ok, err := govalidator.ValidateStruct(_physician)
	if !ok {
		errMsg := "Validate Physician struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Physician struct error: " + err.Error()
		}
		log.Println(errMsg)
		return 0, errors.New(errMsg)
	}
	t := time.Now()
	_physician.CreatedAt = t
	_physician.UpdatedAt = t
	sql := `INSERT INTO physicians (name,created_at,updated_at,introduction) VALUES (:name,:created_at,:updated_at,:introduction)`
	result, err := DB.NamedExec(sql, _physician)
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

// AppointmentsCreate is used for Physician to create the associated objects Appointments
func (_physician *Physician) AppointmentsCreate(am map[string]interface{}) error {
	am["physician_id"] = _physician.Id
	_, err := CreateAppointment(am)
	return err
}

// GetAppointments is used for Physician to get associated objects Appointments
// Say you have a Physician object named physician, when you call physician.GetAppointments(),
// the object will get the associated Appointments attributes evaluated in the struct.
func (_physician *Physician) GetAppointments() error {
	_appointments, err := PhysicianGetAppointments(_physician.Id)
	if err == nil {
		_physician.Appointments = _appointments
	}
	return err
}

// PhysicianGetAppointments a helper fuction used to get associated objects for PhysicianIncludesWhere().
func PhysicianGetAppointments(id int64) ([]Appointment, error) {
	_appointments, err := FindAppointmentsBy("physician_id", id)
	return _appointments, err
}

// PatientsCreate is used for Physician to create the associated objects Patients
func (_physician *Physician) PatientsCreate(am map[string]interface{}) error {
	// FIXME: use transaction to create these associated objects
	patientId, err := CreatePatient(am)
	if err != nil {
		return err
	}
	_, err = CreateAppointment(map[string]interface{}{"physician_id": _physician.Id, "patient_id": patientId})
	return err
}

// GetPatients is used for Physician to get associated objects Patients
// Say you have a Physician object named physician, when you call physician.GetPatients(),
// the object will get the associated Patients attributes evaluated in the struct.
func (_physician *Physician) GetPatients() error {
	_patients, err := PhysicianGetPatients(_physician.Id)
	if err == nil {
		_physician.Patients = _patients
	}
	return err
}

// PhysicianGetPatients a helper fuction used to get associated objects for PhysicianIncludesWhere().
func PhysicianGetPatients(id int64) ([]Patient, error) {
	// FIXME: use transaction to create these associated objects
	sql := `SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at
		        FROM   patients
		               INNER JOIN appointments
		                       ON patients.id = appointments.patient_id
		        WHERE appointments.physician_id = ?`
	_patients, err := FindPatientsBySql(sql, id)
	return _patients, err
}

// PicturesCreate is used for Physician to create the associated objects Pictures
func (_physician *Physician) PicturesCreate(am map[string]interface{}) error {
	am["imageable_id"] = _physician.Id
	am["imageable_type"] = "Physician"
	_, err := CreatePicture(am)
	return err
}

// GetPictures is used for Physician to get associated objects Pictures
// Say you have a Physician object named physician, when you call physician.GetPictures(),
// the object will get the associated Pictures attributes evaluated in the struct.
func (_physician *Physician) GetPictures() error {
	_pictures, err := PhysicianGetPictures(_physician.Id)
	if err == nil {
		_physician.Pictures = _pictures
	}
	return err
}

// PhysicianGetPictures a helper fuction used to get associated objects for PhysicianIncludesWhere().
func PhysicianGetPictures(id int64) ([]Picture, error) {
	where := `imageable_type = "Physician" AND imageable_id = ?`
	_pictures, err := FindPicturesWhere(where, id)
	return _pictures, err
}

// Destroy is method used for a Physician object to be destroyed.
func (_physician *Physician) Destroy() error {
	if _physician.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := DestroyPhysician(_physician.Id)
	return err
}

// DestroyPhysician will destroy a Physician record specified by the id parameter.
func DestroyPhysician(id int64) error {
	stmt, err := DB.Preparex(DB.Rebind(`DELETE FROM physicians WHERE id = ?`))
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

// DestroyPhysicians will destroy Physician records those specified by the ids parameters.
func DestroyPhysicians(ids ...int64) (int64, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return 0, errors.New(msg)
	}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := fmt.Sprintf(`DELETE FROM physicians WHERE id IN (?%s)`, idsHolder)
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

// DestroyPhysiciansWhere delete records by a where clause restriction.
// e.g. DestroyPhysiciansWhere("name = ?", "John")
// And this func will not call the association dependent action
func DestroyPhysiciansWhere(where string, args ...interface{}) (int64, error) {
	sql := `DELETE FROM physicians WHERE `
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

// Save method is used for a Physician object to update an existed record mainly.
// If no id provided a new record will be created. FIXME: A UPSERT action will be implemented further.
func (_physician *Physician) Save() error {
	ok, err := govalidator.ValidateStruct(_physician)
	if !ok {
		errMsg := "Validate Physician struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Physician struct error: " + err.Error()
		}
		log.Println(errMsg)
		return errors.New(errMsg)
	}
	if _physician.Id == 0 {
		_, err = _physician.Create()
		return err
	}
	_physician.UpdatedAt = time.Now()
	sqlFmt := `UPDATE physicians SET %s WHERE id = %v`
	sqlStr := fmt.Sprintf(sqlFmt, "name = :name, updated_at = :updated_at, introduction = :introduction", _physician.Id)
	_, err = DB.NamedExec(sqlStr, _physician)
	return err
}

// UpdatePhysician is used to update a record with a id and map[string]interface{} typed key-value parameters.
func UpdatePhysician(id int64, am map[string]interface{}) error {
	if len(am) == 0 {
		return errors.New("Zero key in the attributes map!")
	}
	am["updated_at"] = time.Now()
	keys := allKeys(am)
	sqlFmt := `UPDATE physicians SET %s WHERE id = %v`
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

// Update is a method used to update a Physician record with the map[string]interface{} typed key-value parameters.
func (_physician *Physician) Update(am map[string]interface{}) error {
	if _physician.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePhysician(_physician.Id, am)
	return err
}

// UpdateAttributes method is supposed to be used to update Physician records as corresponding update_attributes in Ruby on Rails.
func (_physician *Physician) UpdateAttributes(am map[string]interface{}) error {
	if _physician.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePhysician(_physician.Id, am)
	return err
}

// UpdateColumns method is supposed to be used to update Physician records as corresponding update_columns in Ruby on Rails.
func (_physician *Physician) UpdateColumns(am map[string]interface{}) error {
	if _physician.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePhysician(_physician.Id, am)
	return err
}

// UpdatePhysiciansBySql is used to update Physician records by a SQL clause
// using the '?' binding syntax.
func UpdatePhysiciansBySql(sql string, args ...interface{}) (int64, error) {
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
