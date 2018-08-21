// Package models includes the functions on the model Appointment.
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

type Appointment struct {
	Id              int64     `json:"id,omitempty" db:"id" valid:"-"`
	AppointmentDate time.Time `json:"appointment_date,omitempty" db:"appointment_date" valid:"-"`
	PhysicianId     int64     `json:"physician_id,omitempty" db:"physician_id" valid:"-"`
	PatientId       int64     `json:"patient_id,omitempty" db:"patient_id" valid:"-"`
	CreatedAt       time.Time `json:"created_at,omitempty" db:"created_at" valid:"-"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" db:"updated_at" valid:"-"`
	Physician       Physician `json:"physician,omitempty" db:"physician" valid:"-"`
	Patient         Patient   `json:"patient,omitempty" db:"patient" valid:"-"`
}

// DataStruct for the pagination
type AppointmentPage struct {
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

// Current get the current page of AppointmentPage object for pagination.
func (_p *AppointmentPage) Current() ([]Appointment, error) {
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
	appointments, err := FindAppointmentsWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(appointments) != 0 {
		_p.FirstId, _p.LastId = appointments[0].Id, appointments[len(appointments)-1].Id
	}
	return appointments, nil
}

// Previous get the previous page of AppointmentPage object for pagination.
func (_p *AppointmentPage) Previous() ([]Appointment, error) {
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
	appointments, err := FindAppointmentsWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(appointments) != 0 {
		_p.FirstId, _p.LastId = appointments[0].Id, appointments[len(appointments)-1].Id
	}
	_p.PageNum -= 1
	return appointments, nil
}

// Next get the next page of AppointmentPage object for pagination.
func (_p *AppointmentPage) Next() ([]Appointment, error) {
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
	appointments, err := FindAppointmentsWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(appointments) != 0 {
		_p.FirstId, _p.LastId = appointments[0].Id, appointments[len(appointments)-1].Id
	}
	_p.PageNum += 1
	return appointments, nil
}

// GetPage is a helper function for the AppointmentPage object to return a corresponding page due to
// the parameter passed in, i.e. one of "previous, current or next".
func (_p *AppointmentPage) GetPage(direction string) (ps []Appointment, err error) {
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

// buildOrder is for AppointmentPage object to build a SQL ORDER BY clause.
func (_p *AppointmentPage) buildOrder() {
	tempList := []string{}
	for k, v := range _p.Order {
		tempList = append(tempList, fmt.Sprintf("%v %v", k, v))
	}
	_p.orderStr = " ORDER BY " + strings.Join(tempList, ", ")
}

// buildIdRestrict is for AppointmentPage object to build a SQL clause for ID restriction,
// implementing a simple keyset style pagination.
func (_p *AppointmentPage) buildIdRestrict(direction string) (idStr string, idParams []interface{}) {
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

// buildPageCount calculate the TotalItems/TotalPages for the AppointmentPage object.
func (_p *AppointmentPage) buildPageCount() error {
	count, err := AppointmentCountWhere(_p.WhereString, _p.WhereParams...)
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

// FindAppointment find a single appointment by an ID.
func FindAppointment(id int64) (*Appointment, error) {
	if id == 0 {
		return nil, errors.New("Invalid ID: it can't be zero")
	}
	_appointment := Appointment{}
	err := DB.Get(&_appointment, DB.Rebind(`SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments WHERE appointments.id = ? LIMIT 1`), id)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_appointment, nil
}

// FirstAppointment find the first one appointment by ID ASC order.
func FirstAppointment() (*Appointment, error) {
	_appointment := Appointment{}
	err := DB.Get(&_appointment, DB.Rebind(`SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments ORDER BY appointments.id ASC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_appointment, nil
}

// FirstAppointments find the first N appointments by ID ASC order.
func FirstAppointments(n uint32) ([]Appointment, error) {
	_appointments := []Appointment{}
	sql := fmt.Sprintf("SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments ORDER BY appointments.id ASC LIMIT %v", n)
	err := DB.Select(&_appointments, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _appointments, nil
}

// LastAppointment find the last one appointment by ID DESC order.
func LastAppointment() (*Appointment, error) {
	_appointment := Appointment{}
	err := DB.Get(&_appointment, DB.Rebind(`SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments ORDER BY appointments.id DESC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_appointment, nil
}

// LastAppointments find the last N appointments by ID DESC order.
func LastAppointments(n uint32) ([]Appointment, error) {
	_appointments := []Appointment{}
	sql := fmt.Sprintf("SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments ORDER BY appointments.id DESC LIMIT %v", n)
	err := DB.Select(&_appointments, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _appointments, nil
}

// FindAppointments find one or more appointments by the given ID(s).
func FindAppointments(ids ...int64) ([]Appointment, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return nil, errors.New(msg)
	}
	_appointments := []Appointment{}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := DB.Rebind(fmt.Sprintf(`SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments WHERE appointments.id IN (?%s)`, idsHolder))
	idsT := []interface{}{}
	for _, id := range ids {
		idsT = append(idsT, interface{}(id))
	}
	err := DB.Select(&_appointments, sql, idsT...)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _appointments, nil
}

// FindAppointmentBy find a single appointment by a field name and a value.
func FindAppointmentBy(field string, val interface{}) (*Appointment, error) {
	_appointment := Appointment{}
	sqlFmt := `SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments WHERE %s = ? LIMIT 1`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err := DB.Get(&_appointment, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_appointment, nil
}

// FindAppointmentsBy find all appointments by a field name and a value.
func FindAppointmentsBy(field string, val interface{}) (_appointments []Appointment, err error) {
	sqlFmt := `SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments WHERE %s = ?`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err = DB.Select(&_appointments, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _appointments, nil
}

// AllAppointments get all the Appointment records.
func AllAppointments() (appointments []Appointment, err error) {
	err = DB.Select(&appointments, "SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return appointments, nil
}

// AppointmentCount get the count of all the Appointment records.
func AppointmentCount() (c int64, err error) {
	err = DB.Get(&c, "SELECT count(*) FROM appointments")
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return c, nil
}

// AppointmentCountWhere get the count of all the Appointment records with a where clause.
func AppointmentCountWhere(where string, args ...interface{}) (c int64, err error) {
	sql := "SELECT count(*) FROM appointments"
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

// AppointmentIncludesWhere get the Appointment associated models records, currently it's not same as the corresponding "includes" function but "preload" instead in Ruby on Rails. It means that the "sql" should be restricted on Appointment model.
func AppointmentIncludesWhere(assocs []string, sql string, args ...interface{}) (_appointments []Appointment, err error) {
	_appointments, err = FindAppointmentsWhere(sql, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(assocs) == 0 {
		log.Println("No associated fields ard specified")
		return _appointments, err
	}
	if len(_appointments) <= 0 {
		return nil, errors.New("No results available")
	}
	ids := make([]interface{}, len(_appointments))
	for _, v := range _appointments {
		ids = append(ids, interface{}(v.Id))
	}
	return _appointments, nil
}

// AppointmentIds get all the IDs of Appointment records.
func AppointmentIds() (ids []int64, err error) {
	err = DB.Select(&ids, "SELECT id FROM appointments")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return ids, nil
}

// AppointmentIdsWhere get all the IDs of Appointment records by where restriction.
func AppointmentIdsWhere(where string, args ...interface{}) ([]int64, error) {
	ids, err := AppointmentIntCol("id", where, args...)
	return ids, err
}

// AppointmentIntCol get some int64 typed column of Appointment by where restriction.
func AppointmentIntCol(col, where string, args ...interface{}) (intColRecs []int64, err error) {
	sql := "SELECT " + col + " FROM appointments"
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

// AppointmentStrCol get some string typed column of Appointment by where restriction.
func AppointmentStrCol(col, where string, args ...interface{}) (strColRecs []string, err error) {
	sql := "SELECT " + col + " FROM appointments"
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

// FindAppointmentsWhere query use a partial SQL clause that usually following after WHERE
// with placeholders, eg: FindUsersWhere("first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindAppointmentsWhere(where string, args ...interface{}) (appointments []Appointment, err error) {
	sql := "SELECT COALESCE(appointments.appointment_date, CONVERT_TZ('0001-01-01 00:00:00','+00:00','UTC')) AS appointment_date, COALESCE(appointments.physician_id, 0) AS physician_id, COALESCE(appointments.patient_id, 0) AS patient_id, appointments.id, appointments.created_at, appointments.updated_at FROM appointments"
	if len(where) > 0 {
		sql = sql + " WHERE " + where
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&appointments, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return appointments, nil
}

// FindAppointmentBySql query use a complete SQL clause
// with placeholders, eg: FindUserBySql("SELECT * FROM users WHERE first_name = ? AND age > ? ORDER BY DESC LIMIT 1", "John", 18)
// will return only One record in the table "users" whose first_name is "John" and age elder than 18.
func FindAppointmentBySql(sql string, args ...interface{}) (*Appointment, error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_appointment := &Appointment{}
	err = stmt.Get(_appointment, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return _appointment, nil
}

// FindAppointmentsBySql query use a complete SQL clause
// with placeholders, eg: FindUsersBySql("SELECT * FROM users WHERE first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindAppointmentsBySql(sql string, args ...interface{}) (appointments []Appointment, err error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&appointments, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return appointments, nil
}

// CreateAppointment use a named params to create a single Appointment record.
// A named params is key-value map like map[string]interface{}{"first_name": "John", "age": 23} .
func CreateAppointment(am map[string]interface{}) (int64, error) {
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
	sqlFmt := `INSERT INTO appointments (%s) VALUES (%s)`
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

// Create is a method for Appointment to create a record.
func (_appointment *Appointment) Create() (int64, error) {
	ok, err := govalidator.ValidateStruct(_appointment)
	if !ok {
		errMsg := "Validate Appointment struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Appointment struct error: " + err.Error()
		}
		log.Println(errMsg)
		return 0, errors.New(errMsg)
	}
	t := time.Now()
	_appointment.CreatedAt = t
	_appointment.UpdatedAt = t
	sql := `INSERT INTO appointments (appointment_date,physician_id,patient_id,created_at,updated_at) VALUES (:appointment_date,:physician_id,:patient_id,:created_at,:updated_at)`
	result, err := DB.NamedExec(sql, _appointment)
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

// CreatePhysician is a method for a Appointment object to create an associated Physician record.
func (_appointment *Appointment) CreatePhysician(am map[string]interface{}) error {
	am["appointment_id"] = _appointment.Id
	_, err := CreatePhysician(am)
	return err
}

// CreatePatient is a method for a Appointment object to create an associated Patient record.
func (_appointment *Appointment) CreatePatient(am map[string]interface{}) error {
	am["appointment_id"] = _appointment.Id
	_, err := CreatePatient(am)
	return err
}

// Destroy is method used for a Appointment object to be destroyed.
func (_appointment *Appointment) Destroy() error {
	if _appointment.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := DestroyAppointment(_appointment.Id)
	return err
}

// DestroyAppointment will destroy a Appointment record specified by the id parameter.
func DestroyAppointment(id int64) error {
	stmt, err := DB.Preparex(DB.Rebind(`DELETE FROM appointments WHERE id = ?`))
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

// DestroyAppointments will destroy Appointment records those specified by the ids parameters.
func DestroyAppointments(ids ...int64) (int64, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return 0, errors.New(msg)
	}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := fmt.Sprintf(`DELETE FROM appointments WHERE id IN (?%s)`, idsHolder)
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

// DestroyAppointmentsWhere delete records by a where clause restriction.
// e.g. DestroyAppointmentsWhere("name = ?", "John")
// And this func will not call the association dependent action
func DestroyAppointmentsWhere(where string, args ...interface{}) (int64, error) {
	sql := `DELETE FROM appointments WHERE `
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

// Save method is used for a Appointment object to update an existed record mainly.
// If no id provided a new record will be created. FIXME: A UPSERT action will be implemented further.
func (_appointment *Appointment) Save() error {
	ok, err := govalidator.ValidateStruct(_appointment)
	if !ok {
		errMsg := "Validate Appointment struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Appointment struct error: " + err.Error()
		}
		log.Println(errMsg)
		return errors.New(errMsg)
	}
	if _appointment.Id == 0 {
		_, err = _appointment.Create()
		return err
	}
	_appointment.UpdatedAt = time.Now()
	sqlFmt := `UPDATE appointments SET %s WHERE id = %v`
	sqlStr := fmt.Sprintf(sqlFmt, "appointment_date = :appointment_date, physician_id = :physician_id, patient_id = :patient_id, updated_at = :updated_at", _appointment.Id)
	_, err = DB.NamedExec(sqlStr, _appointment)
	return err
}

// UpdateAppointment is used to update a record with a id and map[string]interface{} typed key-value parameters.
func UpdateAppointment(id int64, am map[string]interface{}) error {
	if len(am) == 0 {
		return errors.New("Zero key in the attributes map!")
	}
	am["updated_at"] = time.Now()
	keys := allKeys(am)
	sqlFmt := `UPDATE appointments SET %s WHERE id = %v`
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

// Update is a method used to update a Appointment record with the map[string]interface{} typed key-value parameters.
func (_appointment *Appointment) Update(am map[string]interface{}) error {
	if _appointment.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdateAppointment(_appointment.Id, am)
	return err
}

// UpdateAttributes method is supposed to be used to update Appointment records as corresponding update_attributes in Ruby on Rails.
func (_appointment *Appointment) UpdateAttributes(am map[string]interface{}) error {
	if _appointment.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdateAppointment(_appointment.Id, am)
	return err
}

// UpdateColumns method is supposed to be used to update Appointment records as corresponding update_columns in Ruby on Rails.
func (_appointment *Appointment) UpdateColumns(am map[string]interface{}) error {
	if _appointment.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdateAppointment(_appointment.Id, am)
	return err
}

// UpdateAppointmentsBySql is used to update Appointment records by a SQL clause
// using the '?' binding syntax.
func UpdateAppointmentsBySql(sql string, args ...interface{}) (int64, error) {
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
