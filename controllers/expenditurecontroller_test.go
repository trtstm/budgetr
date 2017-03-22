package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/trtstm/budgetr/db"
	"github.com/trtstm/budgetr/models"

	. "github.com/smartystreets/goconvey/convey"
)

func withDb(cb func()) {
	db.Shutdown()

	if err := db.SetupConnection(db.SQLITE, "file:memdb1?mode=memory&cache=shared"); err != nil {
		panic(err)
	}

	if err := db.SetupSchema(); err != nil {
		panic(err)
	}

	cb()

	if err := db.Shutdown(); err != nil {
		panic(err)
	}
}

type expenditureListResponse struct {
	Data   []*ExpenditureResponse `json:"data"`
	Limit  uint                   `json:"limit"`
	Offset uint                   `json:"offset"`
}

func TestExpenditureControllerIndex(t *testing.T) {
	e := echo.New()

	withDb(func() {
		expenditures := map[uint]*models.Expenditure{}
		Convey("Inserting expenditures.", t, func() {
			var expenditure *models.Expenditure

			expenditure = &models.Expenditure{
				Amount: 123,
				Date:   time.Now(),
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   -1,
				Date:     time.Now().Add(24 * time.Hour),
				Category: &models.Category{Name: "cat1"},
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   -100.53,
				Date:     time.Now().Add(48 * time.Hour),
				Category: &models.Category{Name: "cat2"},
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   100.53,
				Date:     time.Now().Add(100 * time.Hour),
				Category: expenditure.Category, // Use same category as previous.
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure
		})

		Convey("Check if returned data is correct.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			So(len(answer.Data), ShouldEqual, len(expenditures))
			for _, expresp := range answer.Data {
				exp := expenditures[expresp.ID]
				So(expresp.ID, ShouldEqual, exp.ID)
				So(expresp.Amount, ShouldEqual, exp.Amount)
				So(expresp.Date.String(), ShouldEqual, exp.Date.String())

				if exp.Category != nil {
					So(expresp.Category, ShouldNotBeNil)
					So(expresp.Category.ID, ShouldEqual, exp.Category.ID)
					So(expresp.Category.Name, ShouldEqual, exp.Category.Name)
				}
			}
		})

		Convey("Sorting on ascending id.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures?sort=id|asc", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			for i, expresp := range answer.Data {
				if i == 0 {
					continue
				}

				So(expresp.ID, ShouldBeGreaterThan, answer.Data[i-1].ID)
			}
		})

		Convey("Sorting on descending id.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures?sort=id|DESC", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			for i, expresp := range answer.Data {
				if i == 0 {
					continue
				}

				So(expresp.ID, ShouldBeLessThan, answer.Data[i-1].ID)
			}
		})

		Convey("Sorting on descending amount.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures?sort=amount|desc", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			for i, expresp := range answer.Data {
				if i == 0 {
					continue
				}

				So(expresp.Amount, ShouldBeLessThanOrEqualTo, answer.Data[i-1].Amount)
			}
		})

		Convey("Sorting on ascending amount.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures?sort=amount|asc", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			for i, expresp := range answer.Data {
				if i == 0 {
					continue
				}

				So(expresp.Amount, ShouldBeGreaterThanOrEqualTo, answer.Data[i-1].Amount)
			}
		})

		Convey("Sorting on ascending date.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures?sort=date|ASC", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			for i, expresp := range answer.Data {
				if i == 0 {
					continue
				}

				So(expresp.Date.Sub(answer.Data[i-1].Date), ShouldBeGreaterThanOrEqualTo, 0)
			}
		})

		Convey("Sorting on descending date.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures?sort=DATE|desc", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			for i, expresp := range answer.Data {
				if i == 0 {
					continue
				}

				So(expresp.Date.Sub(answer.Data[i-1].Date), ShouldBeLessThanOrEqualTo, 0)
			}
		})

		// passed in limit -> [expected answer.limit, expected len(answer.data)]
		limitTests := map[int][2]int{
			3:     [2]int{3, 3},
			123:   [2]int{100, 4},
			10000: [2]int{100, 4},
		}

		for limit, expected := range limitTests {
			Convey("Checking limit "+strconv.Itoa(limit)+".", t, func() {
				r := httptest.NewRequest("GET", "/api/expenditures?limit="+strconv.Itoa(limit), nil)
				w := httptest.NewRecorder()
				c := e.NewContext(r, w)
				So(ExpenditureController.Index(c), ShouldBeNil)
				So(w.Code, ShouldEqual, http.StatusOK)

				answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
				So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
				So(answer.Limit, ShouldEqual, expected[0])
				So(len(answer.Data), ShouldEqual, expected[1])
			})
		}

		// passed in offset -> [expected answer.offset, expected len(answer.data)]
		offsetTests := map[int][2]int{
			3:     [2]int{3, 1},
			123:   [2]int{123, 0},
			10000: [2]int{10000, 0},
		}

		for offset, expected := range offsetTests {
			Convey("Checking offset "+strconv.Itoa(offset)+".", t, func() {
				r := httptest.NewRequest("GET", "/api/expenditures?offset="+strconv.Itoa(offset), nil)
				w := httptest.NewRecorder()
				c := e.NewContext(r, w)
				So(ExpenditureController.Index(c), ShouldBeNil)
				So(w.Code, ShouldEqual, http.StatusOK)

				answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
				So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
				So(answer.Offset, ShouldEqual, expected[0])
				So(len(answer.Data), ShouldEqual, expected[1])
			})
		}

		Convey("Checking limit + offset.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures?limit=3&offset=2", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			So(ExpenditureController.Index(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)

			answer := &expenditureListResponse{Data: []*ExpenditureResponse{}}
			So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
			So(answer.Offset, ShouldEqual, 2)
			So(answer.Limit, ShouldEqual, 3)
			So(len(answer.Data), ShouldEqual, 2)
		})

	})
}

func TestExpenditureControllerShow(t *testing.T) {
	e := echo.New()

	withDb(func() {
		expenditures := map[uint]*models.Expenditure{}
		Convey("Inserting expenditures.", t, func() {
			var expenditure *models.Expenditure

			expenditure = &models.Expenditure{
				Amount: 10.12,
				Date:   time.Now(),
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   -1,
				Date:     time.Now().Add(24 * time.Hour),
				Category: &models.Category{Name: "cat1"},
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure
		})

		for _, expenditure := range expenditures {
			Convey("Getting expenditure "+strconv.Itoa(int(expenditure.ID))+".", t, func() {
				r := httptest.NewRequest("GET", "/api/expenditures/"+strconv.Itoa(int(expenditure.ID)), nil)
				w := httptest.NewRecorder()
				c := e.NewContext(r, w)
				c.SetParamNames("id")
				c.SetParamValues(strconv.Itoa(int(expenditure.ID)))
				So(ExpenditureController.Show(c), ShouldBeNil)
				So(w.Code, ShouldEqual, http.StatusOK)

				answer := &ExpenditureResponse{}
				So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
				So(answer.ID, ShouldEqual, expenditure.ID)
				So(answer.Amount, ShouldEqual, expenditure.Amount)
				So(answer.Date.String(), ShouldEqual, expenditure.Date.String())
				if expenditure.Category != nil {
					So(answer.Category, ShouldNotBeNil)
					So(answer.Category.ID, ShouldEqual, expenditure.Category.ID)
					So(answer.Category.Name, ShouldEqual, expenditure.Category.Name)
				}
			})
		}

		Convey("Getting non existing expenditure.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures/213", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("213")
			So(ExpenditureController.Show(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Getting with invalid id.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures/helloworld", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("helloworld")
			So(ExpenditureController.Show(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestExpenditureControllerCreate(t *testing.T) {
	e := echo.New()

	withDb(func() {
		now := time.Now()
		type testData struct {
			data     string
			expected *ExpenditureResponse
		}

		tests := []testData{
			testData{
				`{"amount": 12.5, "date": "` + now.Format(time.RFC3339) + `"}`,
				&ExpenditureResponse{
					ID: 1, Amount: 12.5, Date: now,
				}},
			testData{
				`{"id": 213, "amount": -12.5, "date": "` + now.Format(time.RFC3339) + `", "category": "     cat1  "}`,
				&ExpenditureResponse{
					ID: 2, Amount: -12.5, Date: now, Category: &CategoryResponse{ID: 1, Name: "cat1"},
				}},
			testData{`{"amount": 0, "date": "` + now.Format(time.RFC3339) + `", "category": "cat2"}`,
				&ExpenditureResponse{
					ID: 3, Amount: 0, Date: now, Category: &CategoryResponse{ID: 2, Name: "cat2"},
				}},
			testData{`{"id": 0, "amount": 123, "date": "` + now.Format(time.RFC3339) + `", "category": "cat1"}`,
				&ExpenditureResponse{
					ID: 4, Amount: 123, Date: now, Category: &CategoryResponse{ID: 1, Name: "cat1"},
				}},
		}

		for _, testData := range tests {
			Convey("Creating expenditure through json.", t, func() {
				r := httptest.NewRequest("POST", "/api/expenditures", bytes.NewReader([]byte(testData.data)))
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				c := e.NewContext(r, w)
				So(ExpenditureController.Create(c), ShouldBeNil)
				So(w.Code, ShouldEqual, http.StatusCreated)

				answer := &ExpenditureResponse{}
				So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
				So(answer.ID, ShouldEqual, testData.expected.ID)
				So(answer.Amount, ShouldEqual, testData.expected.Amount)
				So(answer.Date.Format(time.RFC3339), ShouldEqual, testData.expected.Date.Format(time.RFC3339))

				So(answer.Category == nil && testData.expected.Category == nil || answer.Category != nil && testData.expected.Category != nil, ShouldBeTrue)
				if answer.Category != nil {
					So(answer.Category.ID, ShouldEqual, testData.expected.Category.ID)
					So(answer.Category.Name, ShouldEqual, testData.expected.Category.Name)
				}
			})
		}

		tests = []testData{
			testData{
				`{"amount": -12.5, "date": "` + now.Format(time.RFC3339) + `", "category": ""}`,
				&ExpenditureResponse{
					ID: 5, Amount: -12.5, Date: now,
				}},
			testData{
				`{"amount": -12.5, "date": "` + now.Format(time.RFC3339) + `", "category": "    "}`,
				&ExpenditureResponse{
					ID: 6, Amount: -12.5, Date: now,
				}},
		}

		for _, testData := range tests {
			Convey("Creating expenditure through json with empty category.", t, func() {
				r := httptest.NewRequest("POST", "/api/expenditures", bytes.NewReader([]byte(testData.data)))
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				c := e.NewContext(r, w)
				So(ExpenditureController.Create(c), ShouldBeNil)
				So(w.Code, ShouldEqual, http.StatusCreated)

				answer := &ExpenditureResponse{}
				So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
				So(answer.ID, ShouldEqual, testData.expected.ID)
				So(answer.Amount, ShouldEqual, testData.expected.Amount)
				So(answer.Date.Format(time.RFC3339), ShouldEqual, testData.expected.Date.Format(time.RFC3339))
				So(answer.Category, ShouldBeNil)
			})
		}
	})
}

func TestExpenditureControllerUpdate(t *testing.T) {
	e := echo.New()

	withDb(func() {
		expenditures := map[uint]*models.Expenditure{}
		Convey("Inserting expenditures.", t, func() {
			var expenditure *models.Expenditure

			expenditure = &models.Expenditure{
				Amount: 123,
				Date:   time.Now(),
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount: -1,
				Date:   time.Now().Add(24 * time.Hour),
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   -100.53,
				Date:     time.Now().Add(48 * time.Hour),
				Category: &models.Category{Name: "cat2"},
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   100.53,
				Date:     time.Now().Add(100 * time.Hour),
				Category: expenditure.Category, // Use same category as previous.
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure
		})

		type testData struct {
			data     string
			expected *ExpenditureResponse
		}

		someData := time.Now().Add(-1024 * time.Hour)
		tests := []testData{
			testData{
				`{"amount": 12.5, "date": "` + someData.Format(time.RFC3339) + `"}`,
				&ExpenditureResponse{
					ID: 1, Amount: 12.5, Date: someData,
				}},
			testData{
				`{"id": 213, "amount": -12.5, "date": "` + someData.Format(time.RFC3339) + `", "category": "     cat1  "}`,
				&ExpenditureResponse{
					ID: 2, Amount: -12.5, Date: someData, Category: &CategoryResponse{ID: 2, Name: "cat1"},
				}},
			testData{`{"amount": 0, "date": "` + someData.Format(time.RFC3339) + `", "category": "cat2"}`,
				&ExpenditureResponse{
					ID: 3, Amount: 0, Date: someData, Category: &CategoryResponse{ID: 1, Name: "cat2"},
				}},
			testData{`{"id": 0, "amount": 123, "date": "` + someData.Format(time.RFC3339) + `", "category": "cat1"}`,
				&ExpenditureResponse{
					ID: 4, Amount: 123, Date: someData, Category: &CategoryResponse{ID: 2, Name: "cat1"},
				}},
		}

		for _, testData := range tests {
			Convey("Updating expenditure date.", t, func() {
				r := httptest.NewRequest("POST", "/api/expenditures/"+strconv.Itoa(int(testData.expected.ID)), bytes.NewReader([]byte(testData.data)))
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				c := e.NewContext(r, w)
				c.SetParamNames("id")
				c.SetParamValues(strconv.Itoa(int(testData.expected.ID)))
				So(ExpenditureController.Update(c), ShouldBeNil)
				So(w.Code, ShouldEqual, http.StatusOK)

				answer := &ExpenditureResponse{}
				So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
				So(answer.ID, ShouldEqual, testData.expected.ID)
				So(answer.Amount, ShouldEqual, testData.expected.Amount)
				So(answer.Date.Format(time.RFC3339), ShouldEqual, testData.expected.Date.Format(time.RFC3339))
				So(answer.Category == nil && testData.expected.Category == nil || answer.Category != nil && testData.expected.Category != nil, ShouldBeTrue)
				if answer.Category != nil {
					So(answer.Category.ID, ShouldEqual, testData.expected.Category.ID)
					So(answer.Category.Name, ShouldEqual, testData.expected.Category.Name)
				}
			})
		}

		Convey("Updating expenditure 1 again with same values.", t, func() {
			r := httptest.NewRequest("POST", "/api/expenditures/1", bytes.NewReader([]byte(tests[0].data)))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("1")
			So(ExpenditureController.Update(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Updating non existing expenditure.", t, func() {
			r := httptest.NewRequest("POST", "/api/expenditures/123141", bytes.NewReader([]byte(`{"id": 213, "amount": -12.5, "date": "`+someData.Format(time.RFC3339)+`", "category": "     cat1  "}`)))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("123141")
			So(ExpenditureController.Update(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestExpenditureControllerDelete(t *testing.T) {
	e := echo.New()

	withDb(func() {
		expenditures := map[uint]*models.Expenditure{}
		Convey("Inserting expenditures.", t, func() {
			var expenditure *models.Expenditure

			expenditure = &models.Expenditure{
				Amount: 123,
				Date:   time.Now(),
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount: -1,
				Date:   time.Now().Add(24 * time.Hour),
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   -100.53,
				Date:     time.Now().Add(48 * time.Hour),
				Category: &models.Category{Name: "cat2"},
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure

			expenditure = &models.Expenditure{
				Amount:   100.53,
				Date:     time.Now().Add(100 * time.Hour),
				Category: expenditure.Category, // Use same category as previous.
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure
		})

		Convey("Deleting expenditure 1.", t, func() {
			r := httptest.NewRequest("POST", "/api/expenditures/1", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("1")
			So(ExpenditureController.Delete(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Checking if expenditure 1 was deleted.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures/1", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("1")
			So(ExpenditureController.Show(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Deleting expenditure again 1.", t, func() {
			r := httptest.NewRequest("POST", "/api/expenditures/1", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("1")
			So(ExpenditureController.Delete(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Deleting expenditure 4.", t, func() {
			r := httptest.NewRequest("POST", "/api/expenditures/4", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("4")
			So(ExpenditureController.Delete(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Checking if expenditure 4 was deleted.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures/4", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("4")
			So(ExpenditureController.Show(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Deleting expenditure again 4.", t, func() {
			r := httptest.NewRequest("POST", "/api/expenditures/4", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("4")
			So(ExpenditureController.Delete(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Checking if expenditure 2 is still here.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures/2", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("2")
			So(ExpenditureController.Show(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Checking if expenditure 3 is still here.", t, func() {
			r := httptest.NewRequest("GET", "/api/expenditures/3", nil)
			w := httptest.NewRecorder()
			c := e.NewContext(r, w)
			c.SetParamNames("id")
			c.SetParamValues("3")
			So(ExpenditureController.Show(c), ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

	})
}
