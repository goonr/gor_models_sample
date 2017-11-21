// Package models includes the functions on the model Patient.
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

type Patient struct {
	Id           int64         `json:"id,omitempty" db:"id" valid:"-"`
	Name         string        `json:"name,omitempty" db:"name" valid:"-"`
	CreatedAt    time.Time     `json:"created_at,omitempty" db:"created_at" valid:"-"`
	UpdatedAt    time.Time     `json:"updated_at,omitempty" db:"updated_at" valid:"-"`
	Appointments []Appointment `json:"appointments,omitempty" db:"appointments" valid:"-"`
	Physicians   []Physician   `json:"physicians,omitempty" db:"physicians" valid:"-"`
}

// DataStruct for the pagination
type PatientPage struct {
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

// Current get the current page of PatientPage object for pagination.
func (_p *PatientPage) Current() ([]Patient, error) {
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
	patients, err := FindPatientsWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(patients) != 0 {
		_p.FirstId, _p.LastId = patients[0].Id, patients[len(patients)-1].Id
	}
	return patients, nil
}

// Previous get the previous page of PatientPage object for pagination.
func (_p *PatientPage) Previous() ([]Patient, error) {
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
	patients, err := FindPatientsWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(patients) != 0 {
		_p.FirstId, _p.LastId = patients[0].Id, patients[len(patients)-1].Id
	}
	_p.PageNum -= 1
	return patients, nil
}

// Next get the next page of PatientPage object for pagination.
func (_p *PatientPage) Next() ([]Patient, error) {
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
	patients, err := FindPatientsWhere(whereStr, whereParams...)
	if err != nil {
		return nil, err
	}
	if len(patients) != 0 {
		_p.FirstId, _p.LastId = patients[0].Id, patients[len(patients)-1].Id
	}
	_p.PageNum += 1
	return patients, nil
}

// GetPage is a helper function for the PatientPage object to return a corresponding page due to
// the parameter passed in, i.e. one of "previous, current or next".
func (_p *PatientPage) GetPage(direction string) (ps []Patient, err error) {
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

// buildOrder is for PatientPage object to build a SQL ORDER BY clause.
func (_p *PatientPage) buildOrder() {
	tempList := []string{}
	for k, v := range _p.Order {
		tempList = append(tempList, fmt.Sprintf("%v %v", k, v))
	}
	_p.orderStr = " ORDER BY " + strings.Join(tempList, ", ")
}

// buildIdRestrict is for PatientPage object to build a SQL clause for ID restriction,
// implementing a simple keyset style pagination.
func (_p *PatientPage) buildIdRestrict(direction string) (idStr string, idParams []interface{}) {
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

// buildPageCount calculate the TotalItems/TotalPages for the PatientPage object.
func (_p *PatientPage) buildPageCount() error {
	count, err := PatientCountWhere(_p.WhereString, _p.WhereParams...)
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

// FindPatient find a single patient by an ID.
func FindPatient(id int64) (*Patient, error) {
	if id == 0 {
		return nil, errors.New("Invalid ID: it can't be zero")
	}
	_patient := Patient{}
	err := DB.Get(&_patient, DB.Rebind(`SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients WHERE patients.id = ? LIMIT 1`), id)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_patient, nil
}

// FirstPatient find the first one patient by ID ASC order.
func FirstPatient() (*Patient, error) {
	_patient := Patient{}
	err := DB.Get(&_patient, DB.Rebind(`SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients ORDER BY patients.id ASC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_patient, nil
}

// FirstPatients find the first N patients by ID ASC order.
func FirstPatients(n uint32) ([]Patient, error) {
	_patients := []Patient{}
	sql := fmt.Sprintf("SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients ORDER BY patients.id ASC LIMIT %v", n)
	err := DB.Select(&_patients, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _patients, nil
}

// LastPatient find the last one patient by ID DESC order.
func LastPatient() (*Patient, error) {
	_patient := Patient{}
	err := DB.Get(&_patient, DB.Rebind(`SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients ORDER BY patients.id DESC LIMIT 1`))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_patient, nil
}

// LastPatients find the last N patients by ID DESC order.
func LastPatients(n uint32) ([]Patient, error) {
	_patients := []Patient{}
	sql := fmt.Sprintf("SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients ORDER BY patients.id DESC LIMIT %v", n)
	err := DB.Select(&_patients, DB.Rebind(sql))
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _patients, nil
}

// FindPatients find one or more patients by the given ID(s).
func FindPatients(ids ...int64) ([]Patient, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return nil, errors.New(msg)
	}
	_patients := []Patient{}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := DB.Rebind(fmt.Sprintf(`SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients WHERE patients.id IN (?%s)`, idsHolder))
	idsT := []interface{}{}
	for _, id := range ids {
		idsT = append(idsT, interface{}(id))
	}
	err := DB.Select(&_patients, sql, idsT...)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _patients, nil
}

// FindPatientBy find a single patient by a field name and a value.
func FindPatientBy(field string, val interface{}) (*Patient, error) {
	_patient := Patient{}
	sqlFmt := `SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients WHERE %s = ? LIMIT 1`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err := DB.Get(&_patient, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return &_patient, nil
}

// FindPatientsBy find all patients by a field name and a value.
func FindPatientsBy(field string, val interface{}) (_patients []Patient, err error) {
	sqlFmt := `SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients WHERE %s = ?`
	sqlStr := fmt.Sprintf(sqlFmt, field)
	err = DB.Select(&_patients, DB.Rebind(sqlStr), val)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}
	return _patients, nil
}

// AllPatients get all the Patient records.
func AllPatients() (patients []Patient, err error) {
	err = DB.Select(&patients, "SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return patients, nil
}

// PatientCount get the count of all the Patient records.
func PatientCount() (c int64, err error) {
	err = DB.Get(&c, "SELECT count(*) FROM patients")
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return c, nil
}

// PatientCountWhere get the count of all the Patient records with a where clause.
func PatientCountWhere(where string, args ...interface{}) (c int64, err error) {
	sql := "SELECT count(*) FROM patients"
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

// PatientIncludesWhere get the Patient associated models records, currently it's not same as the corresponding "includes" function but "preload" instead in Ruby on Rails. It means that the "sql" should be restricted on Patient model.
func PatientIncludesWhere(assocs []string, sql string, args ...interface{}) (_patients []Patient, err error) {
	_patients, err = FindPatientsWhere(sql, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(assocs) == 0 {
		log.Println("No associated fields ard specified")
		return _patients, err
	}
	if len(_patients) <= 0 {
		return nil, errors.New("No results available")
	}
	ids := make([]interface{}, len(_patients))
	for _, v := range _patients {
		ids = append(ids, interface{}(v.Id))
	}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	for _, assoc := range assocs {
		switch assoc {
		case "appointments":
			where := fmt.Sprintf("patient_id IN (?%s)", idsHolder)
			_appointments, err := FindAppointmentsWhere(where, ids...)
			if err != nil {
				log.Printf("Error when query associated objects: %v\n", assoc)
				continue
			}
			for _, vv := range _appointments {
				for i, vvv := range _patients {
					if vv.PatientId == vvv.Id {
						vvv.Appointments = append(vvv.Appointments, vv)
					}
					_patients[i].Appointments = vvv.Appointments
				}
			}
		case "physicians":
			// FIXME: optimize the query
			for i, vvv := range _patients {
				_physicians, err := PatientGetPhysicians(vvv.Id)
				if err != nil {
					continue
				}
				vvv.Physicians = _physicians
				_patients[i] = vvv
			}
		}
	}
	return _patients, nil
}

// PatientIds get all the IDs of Patient records.
func PatientIds() (ids []int64, err error) {
	err = DB.Select(&ids, "SELECT id FROM patients")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return ids, nil
}

// PatientIdsWhere get all the IDs of Patient records by where restriction.
func PatientIdsWhere(where string, args ...interface{}) ([]int64, error) {
	ids, err := PatientIntCol("id", where, args...)
	return ids, err
}

// PatientIntCol get some int64 typed column of Patient by where restriction.
func PatientIntCol(col, where string, args ...interface{}) (intColRecs []int64, err error) {
	sql := "SELECT " + col + " FROM patients"
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

// PatientStrCol get some string typed column of Patient by where restriction.
func PatientStrCol(col, where string, args ...interface{}) (strColRecs []string, err error) {
	sql := "SELECT " + col + " FROM patients"
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

// FindPatientsWhere query use a partial SQL clause that usually following after WHERE
// with placeholders, eg: FindUsersWhere("first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindPatientsWhere(where string, args ...interface{}) (patients []Patient, err error) {
	sql := "SELECT COALESCE(patients.name, '') AS name, patients.id, patients.created_at, patients.updated_at FROM patients"
	if len(where) > 0 {
		sql = sql + " WHERE " + where
	}
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&patients, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return patients, nil
}

// FindPatientBySql query use a complete SQL clause
// with placeholders, eg: FindUserBySql("SELECT * FROM users WHERE first_name = ? AND age > ? ORDER BY DESC LIMIT 1", "John", 18)
// will return only One record in the table "users" whose first_name is "John" and age elder than 18.
func FindPatientBySql(sql string, args ...interface{}) (*Patient, error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_patient := &Patient{}
	err = stmt.Get(_patient, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return _patient, nil
}

// FindPatientsBySql query use a complete SQL clause
// with placeholders, eg: FindUsersBySql("SELECT * FROM users WHERE first_name = ? AND age > ?", "John", 18)
// will return those records in the table "users" whose first_name is "John" and age elder than 18.
func FindPatientsBySql(sql string, args ...interface{}) (patients []Patient, err error) {
	stmt, err := DB.Preparex(DB.Rebind(sql))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = stmt.Select(&patients, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return patients, nil
}

// CreatePatient use a named params to create a single Patient record.
// A named params is key-value map like map[string]interface{}{"first_name": "John", "age": 23} .
func CreatePatient(am map[string]interface{}) (int64, error) {
	if len(am) == 0 {
		return 0, fmt.Errorf("Zero key in the attributes map!")
	}
	t := time.Now()
	for _, v := range []string{"created_at", "updated_at"} {
		if am[v] == nil {
			am[v] = t
		}
	}
	keys := make([]string, len(am))
	i := 0
	for k := range am {
		keys[i] = k
		i++
	}
	sqlFmt := `INSERT INTO patients (%s) VALUES (%s)`
	sqlStr := fmt.Sprintf(sqlFmt, strings.Join(keys, ","), ":"+strings.Join(keys, ",:"))
	result, err := DB.NamedExec(sqlStr, am)
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

// Create is a method for Patient to create a record.
func (_patient *Patient) Create() (int64, error) {
	ok, err := govalidator.ValidateStruct(_patient)
	if !ok {
		errMsg := "Validate Patient struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Patient struct error: " + err.Error()
		}
		log.Println(errMsg)
		return 0, errors.New(errMsg)
	}
	t := time.Now()
	_patient.CreatedAt = t
	_patient.UpdatedAt = t
	sql := `INSERT INTO patients (name,created_at,updated_at) VALUES (:name,:created_at,:updated_at)`
	result, err := DB.NamedExec(sql, _patient)
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

// AppointmentsCreate is used for Patient to create the associated objects Appointments
func (_patient *Patient) AppointmentsCreate(am map[string]interface{}) error {
	am["patient_id"] = _patient.Id
	_, err := CreateAppointment(am)
	return err
}

// GetAppointments is used for Patient to get associated objects Appointments
// Say you have a Patient object named patient, when you call patient.GetAppointments(),
// the object will get the associated Appointments attributes evaluated in the struct.
func (_patient *Patient) GetAppointments() error {
	_appointments, err := PatientGetAppointments(_patient.Id)
	if err == nil {
		_patient.Appointments = _appointments
	}
	return err
}

// PatientGetAppointments a helper fuction used to get associated objects for PatientIncludesWhere().
func PatientGetAppointments(id int64) ([]Appointment, error) {
	_appointments, err := FindAppointmentsBy("patient_id", id)
	return _appointments, err
}

// PhysiciansCreate is used for Patient to create the associated objects Physicians
func (_patient *Patient) PhysiciansCreate(am map[string]interface{}) error {
	// FIXME: use transaction to create these associated objects
	physicianId, err := CreatePhysician(am)
	if err != nil {
		return err
	}
	_, err = CreateAppointment(map[string]interface{}{"patient_id": _patient.Id, "physician_id": physicianId})
	return err
}

// GetPhysicians is used for Patient to get associated objects Physicians
// Say you have a Patient object named patient, when you call patient.GetPhysicians(),
// the object will get the associated Physicians attributes evaluated in the struct.
func (_patient *Patient) GetPhysicians() error {
	_physicians, err := PatientGetPhysicians(_patient.Id)
	if err == nil {
		_patient.Physicians = _physicians
	}
	return err
}

// PatientGetPhysicians a helper fuction used to get associated objects for PatientIncludesWhere().
func PatientGetPhysicians(id int64) ([]Physician, error) {
	// FIXME: use transaction to create these associated objects
	sql := `SELECT COALESCE(physicians.name, '') AS name, COALESCE(physicians.introduction, '') AS introduction, physicians.id, physicians.created_at, physicians.updated_at
		        FROM   physicians
		               INNER JOIN appointments
		                       ON physicians.id = appointments.physician_id
		        WHERE appointments.patient_id = ?`
	_physicians, err := FindPhysiciansBySql(sql, id)
	return _physicians, err
}

// Destroy is method used for a Patient object to be destroyed.
func (_patient *Patient) Destroy() error {
	if _patient.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := DestroyPatient(_patient.Id)
	return err
}

// DestroyPatient will destroy a Patient record specified by the id parameter.
func DestroyPatient(id int64) error {
	stmt, err := DB.Preparex(DB.Rebind(`DELETE FROM patients WHERE id = ?`))
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

// DestroyPatients will destroy Patient records those specified by the ids parameters.
func DestroyPatients(ids ...int64) (int64, error) {
	if len(ids) == 0 {
		msg := "At least one or more ids needed"
		log.Println(msg)
		return 0, errors.New(msg)
	}
	idsHolder := strings.Repeat(",?", len(ids)-1)
	sql := fmt.Sprintf(`DELETE FROM patients WHERE id IN (?%s)`, idsHolder)
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

// DestroyPatientsWhere delete records by a where clause restriction.
// e.g. DestroyPatientsWhere("name = ?", "John")
// And this func will not call the association dependent action
func DestroyPatientsWhere(where string, args ...interface{}) (int64, error) {
	sql := `DELETE FROM patients WHERE `
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

// Save method is used for a Patient object to update an existed record mainly.
// If no id provided a new record will be created. FIXME: A UPSERT action will be implemented further.
func (_patient *Patient) Save() error {
	ok, err := govalidator.ValidateStruct(_patient)
	if !ok {
		errMsg := "Validate Patient struct error: Unknown error"
		if err != nil {
			errMsg = "Validate Patient struct error: " + err.Error()
		}
		log.Println(errMsg)
		return errors.New(errMsg)
	}
	if _patient.Id == 0 {
		_, err = _patient.Create()
		return err
	}
	_patient.UpdatedAt = time.Now()
	sqlFmt := `UPDATE patients SET %s WHERE id = %v`
	sqlStr := fmt.Sprintf(sqlFmt, "name = :name, updated_at = :updated_at", _patient.Id)
	_, err = DB.NamedExec(sqlStr, _patient)
	return err
}

// UpdatePatient is used to update a record with a id and map[string]interface{} typed key-value parameters.
func UpdatePatient(id int64, am map[string]interface{}) error {
	if len(am) == 0 {
		return errors.New("Zero key in the attributes map!")
	}
	am["updated_at"] = time.Now()
	keys := make([]string, len(am))
	i := 0
	for k := range am {
		keys[i] = k
		i++
	}
	sqlFmt := `UPDATE patients SET %s WHERE id = %v`
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

// Update is a method used to update a Patient record with the map[string]interface{} typed key-value parameters.
func (_patient *Patient) Update(am map[string]interface{}) error {
	if _patient.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePatient(_patient.Id, am)
	return err
}

// UpdateAttributes method is supposed to be used to update Patient records as corresponding update_attributes in Ruby on Rails.
func (_patient *Patient) UpdateAttributes(am map[string]interface{}) error {
	if _patient.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePatient(_patient.Id, am)
	return err
}

// UpdateColumns method is supposed to be used to update Patient records as corresponding update_columns in Ruby on Rails.
func (_patient *Patient) UpdateColumns(am map[string]interface{}) error {
	if _patient.Id == 0 {
		return errors.New("Invalid Id field: it can't be a zero value")
	}
	err := UpdatePatient(_patient.Id, am)
	return err
}

// UpdatePatientsBySql is used to update Patient records by a SQL clause
// using the '?' binding syntax.
func UpdatePatientsBySql(sql string, args ...interface{}) (int64, error) {
	if sql == "" {
		return 0, errors.New("A blank SQL clause")
	}
	sql = strings.Replace(strings.ToLower(sql), "set", "set updated_at = ?, ", 1)
	args = append([]interface{}{time.Now()}, args...)
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
