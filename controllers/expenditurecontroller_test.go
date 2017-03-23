package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

type test struct {
	URL         string
	Method      string
	Params      []string
	ParamValues []string

	Endpoint func(echo.Context) error

	ContentType string
	PostData    string

	ExpectedStatusCode              interface{}
	ExpectedStatusCodeComparison    func(interface{}, ...interface{}) string
	ExpectedExpenditureListResponse *expenditureListResponse
	ExpectedExpenditureResponse     *ExpenditureResponse
}

func doTest(test *test) {
	e := echo.New()

	if test.ExpectedStatusCodeComparison == nil {
		test.ExpectedStatusCodeComparison = ShouldEqual
	}

	var r *http.Request
	if test.PostData != "" {
		r = httptest.NewRequest(test.Method, test.URL, bytes.NewReader([]byte(test.PostData)))
	} else {
		r = httptest.NewRequest(test.Method, test.URL, nil)
	}

	if test.ContentType != "" {
		r.Header.Set("Content-Type", test.ContentType)
	}

	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	c.SetParamNames(test.Params...)
	c.SetParamValues(test.ParamValues...)
	So(test.Endpoint(c), ShouldBeNil)
	So(w.Code, test.ExpectedStatusCodeComparison, test.ExpectedStatusCode)

	if test.ExpectedExpenditureResponse != nil {
		answer := &ExpenditureResponse{}
		So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
		IsExpectedExpenditureResponse(answer, test.ExpectedExpenditureResponse)
	}

	if test.ExpectedExpenditureListResponse != nil {
		answer := &expenditureListResponse{}
		So(json.NewDecoder(w.Result().Body).Decode(answer), ShouldBeNil)
		So(answer.Limit, ShouldEqual, test.ExpectedExpenditureListResponse.Limit)
		So(answer.Offset, ShouldEqual, test.ExpectedExpenditureListResponse.Offset)
		So(len(answer.Data), ShouldEqual, len(test.ExpectedExpenditureListResponse.Data))

		for i, exp := range answer.Data {
			IsExpectedExpenditureResponse(exp, test.ExpectedExpenditureListResponse.Data[i])
		}

	}
}

func IsExpectedExpenditureResponse(actual *ExpenditureResponse, expected *ExpenditureResponse) {
	So(actual.ID, ShouldEqual, expected.ID)
	So(actual.Amount, ShouldEqual, expected.Amount)
	So(actual.Date.Format(time.RFC3339), ShouldEqual, expected.Date.Format(time.RFC3339))

	So(actual.Category != nil && expected.Category != nil || actual.Category == nil && expected.Category == nil, ShouldBeTrue)

	if actual.Category != nil {
		So(actual.Category.ID, ShouldEqual, expected.Category.ID)
		So(actual.Category.Name, ShouldEqual, expected.Category.Name)
	}
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
			r := httptest.NewRequest("GET", "/api/expenditures?sort=id-asc", nil)
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
			r := httptest.NewRequest("GET", "/api/expenditures?sort=id-DESC", nil)
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
			r := httptest.NewRequest("GET", "/api/expenditures?sort=amount-desc", nil)
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
			r := httptest.NewRequest("GET", "/api/expenditures?sort=amount-asc", nil)
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
			r := httptest.NewRequest("GET", "/api/expenditures?sort=date-ASC", nil)
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
			r := httptest.NewRequest("GET", "/api/expenditures?sort=DATE-desc", nil)
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

	withDb(func() {
		now := time.Now()

		db.DB.Create(&models.Expenditure{
			Amount: 2,
			Date:   now.Add(-700 * time.Hour),
		})

		db.DB.Create(&models.Expenditure{
			Amount: 2,
			Date:   now.Add(-366 * time.Hour),
		})

		db.DB.Create(&models.Expenditure{
			Amount: 2,
			Date:   now.Add(-100 * time.Hour),
		})

		db.DB.Create(&models.Expenditure{
			Amount: 2,
			Date:   now.Add(0 * time.Hour),
		})

		db.DB.Create(&models.Expenditure{
			Amount: 2,
			Date:   now.Add(100 * time.Hour),
		})

		dateRange := url.Values{}
		dateRange.Set("sort", "date-desc")

		tests := []test{}

		dateRange.Set("start", now.Add(-500*time.Hour).Format(time.RFC3339))
		dateRange.Set("end", now.Add(50*time.Hour).Format(time.RFC3339))
		tests = append(tests, test{
			URL:                "/api/expenditures?" + dateRange.Encode(),
			Method:             "get",
			Endpoint:           ExpenditureController.Index,
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureListResponse: &expenditureListResponse{
				Limit:  100,
				Offset: 0,
				Data: []*ExpenditureResponse{
					&ExpenditureResponse{
						ID:     4,
						Amount: 2,
						Date:   now.Add(0 * time.Hour),
					},
					&ExpenditureResponse{
						ID:     3,
						Amount: 2,
						Date:   now.Add(-100 * time.Hour),
					},
					&ExpenditureResponse{
						ID:     2,
						Amount: 2,
						Date:   now.Add(-366 * time.Hour),
					},
				},
			},
		})

		dateRange.Set("start", now.Add(-5*time.Hour).Format(time.RFC3339))
		dateRange.Set("end", now.Add(5*time.Hour).Format(time.RFC3339))
		tests = append(tests, test{
			URL:                "/api/expenditures?" + dateRange.Encode(),
			Method:             "get",
			Endpoint:           ExpenditureController.Index,
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureListResponse: &expenditureListResponse{
				Limit:  100,
				Offset: 0,
				Data: []*ExpenditureResponse{
					&ExpenditureResponse{
						ID:     4,
						Amount: 2,
						Date:   now.Add(0 * time.Hour),
					},
				},
			},
		})

		// Test if result is in [start, end)
		dateRange.Set("start", now.Add(-700*time.Hour).Format(time.RFC3339))
		dateRange.Set("end", now.Add(100*time.Hour).Format(time.RFC3339))
		tests = append(tests, test{
			URL:                "/api/expenditures?" + dateRange.Encode(),
			Method:             "get",
			Endpoint:           ExpenditureController.Index,
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureListResponse: &expenditureListResponse{
				Limit:  100,
				Offset: 0,
				Data: []*ExpenditureResponse{
					&ExpenditureResponse{
						ID:     4,
						Amount: 2,
						Date:   now.Add(0 * time.Hour),
					},
					&ExpenditureResponse{
						ID:     3,
						Amount: 2,
						Date:   now.Add(-100 * time.Hour),
					},
					&ExpenditureResponse{
						ID:     2,
						Amount: 2,
						Date:   now.Add(-366 * time.Hour),
					},
					&ExpenditureResponse{
						ID:     1,
						Amount: 2,
						Date:   now.Add(-700 * time.Hour),
					},
				},
			},
		})

		dateRange.Set("start", "asda")
		dateRange.Set("end", now.Add(5*time.Hour).Format(time.RFC3339))
		tests = append(tests, test{
			URL:                "/api/expenditures?" + dateRange.Encode(),
			Method:             "get",
			Endpoint:           ExpenditureController.Index,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		dateRange.Set("start", now.Add(-5*time.Hour).Format(time.RFC3339))
		dateRange.Set("end", "asa")
		tests = append(tests, test{
			URL:                "/api/expenditures?" + dateRange.Encode(),
			Method:             "get",
			Endpoint:           ExpenditureController.Index,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		dateRange.Del("start")
		dateRange.Set("end", now.Add(5*time.Hour).Format(time.RFC3339))
		tests = append(tests, test{
			URL:                "/api/expenditures?" + dateRange.Encode(),
			Method:             "get",
			Endpoint:           ExpenditureController.Index,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		Convey("Checking start,end range.", t, func() {
			for _, test := range tests {
				doTest(&test)
			}
		})
	})
}

func TestExpenditureControllerShow(t *testing.T) {
	withDb(func() {
		tests := []test{}
		expenditures := map[uint]*models.Expenditure{}
		Convey("Inserting expenditures.", t, func() {
			var expenditure *models.Expenditure

			expenditure = &models.Expenditure{
				Amount: 10.12,
				Date:   time.Now(),
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure
			tests = append(tests, test{
				URL:                "/api/expenditures:id",
				Method:             "get",
				Params:             []string{"id"},
				ParamValues:        []string{strconv.Itoa(int(expenditure.ID))},
				Endpoint:           ExpenditureController.Show,
				ExpectedStatusCode: http.StatusOK,
				ExpectedExpenditureResponse: &ExpenditureResponse{
					ID:     expenditure.ID,
					Amount: expenditure.Amount,
					Date:   expenditure.Date,
				},
			})

			expenditure = &models.Expenditure{
				Amount:   -1,
				Date:     time.Now().Add(24 * time.Hour),
				Category: &models.Category{Name: "cat1"},
			}
			db.DB.Create(expenditure)
			expenditures[expenditure.ID] = expenditure
			tests = append(tests, test{
				URL:                "/api/expenditures:id",
				Method:             "get",
				Params:             []string{"id"},
				ParamValues:        []string{strconv.Itoa(int(expenditure.ID))},
				Endpoint:           ExpenditureController.Show,
				ExpectedStatusCode: http.StatusOK,
				ExpectedExpenditureResponse: &ExpenditureResponse{
					ID:     expenditure.ID,
					Amount: expenditure.Amount,
					Date:   expenditure.Date,
					Category: &CategoryResponse{
						ID:   expenditure.CategoryID,
						Name: expenditure.Category.Name,
					},
				},
			})
		})

		tests = append(tests, test{
			URL:                "/api/expenditures:id",
			Method:             "get",
			Params:             []string{"id"},
			ParamValues:        []string{"1231"},
			Endpoint:           ExpenditureController.Show,
			ExpectedStatusCode: http.StatusNotFound,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures:id",
			Method:             "get",
			Params:             []string{"id"},
			ParamValues:        []string{"0"},
			Endpoint:           ExpenditureController.Show,
			ExpectedStatusCode: http.StatusNotFound,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures:id",
			Method:             "get",
			Params:             []string{"id"},
			ParamValues:        []string{"-1"},
			Endpoint:           ExpenditureController.Show,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures:id",
			Method:             "get",
			Params:             []string{"id"},
			ParamValues:        []string{"abc"},
			Endpoint:           ExpenditureController.Show,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures:id",
			Method:             "get",
			Params:             []string{"id"},
			ParamValues:        []string{""},
			Endpoint:           ExpenditureController.Show,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		Convey("Testing show expenditures.", t, func() {
			for _, test := range tests {
				doTest(&test)
			}
		})
	})
}

func TestExpenditureControllerCreate(t *testing.T) {

	withDb(func() {
		now := time.Now()
		tests := []test{}
		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": -1.5, "date": "` + now.Format(time.RFC3339) + `"}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     1,
				Amount: -1.5,
				Date:   now,
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": 0, "date": "` + now.Add(24*time.Hour).Format(time.RFC3339) + `"}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     2,
				Amount: 0,
				Date:   now.Add(24 * time.Hour),
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": 123, "date": "` + now.Format(time.RFC3339) + `", "category": "cat1"}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     3,
				Amount: 123,
				Date:   now,
				Category: &CategoryResponse{
					ID:   1,
					Name: "cat1",
				},
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": 123.368, "date": "` + now.Format(time.RFC3339) + `", "category": "  cat2   "}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     4,
				Amount: 123.368,
				Date:   now,
				Category: &CategoryResponse{
					ID:   2,
					Name: "cat2",
				},
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": 123.368, "date": "` + now.Format(time.RFC3339) + `", "category": "  cat1   "}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     5,
				Amount: 123.368,
				Date:   now,
				Category: &CategoryResponse{
					ID:   1,
					Name: "cat1",
				},
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": 123.368, "date": "` + now.Format(time.RFC3339) + `", "category": "cat2"}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     6,
				Amount: 123.368,
				Date:   now,
				Category: &CategoryResponse{
					ID:   2,
					Name: "cat2",
				},
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": 123.368, "category": "cat2"}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     7,
				Amount: 123.368,
				Category: &CategoryResponse{
					ID:   2,
					Name: "cat2",
				},
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     8,
				Amount: 0,
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": "abc", "date": "` + now.Format(time.RFC3339) + `", "category": "cat2"}`,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `{"amount": "abc", "date": "aa", "category": "cat2"}`,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures",
			Method:             "post",
			Endpoint:           ExpenditureController.Create,
			ContentType:        "application/json",
			PostData:           `}{`,
			ExpectedStatusCode: http.StatusBadRequest,
		})

		Convey("Creating expenditures.", t, func() {
			for _, test := range tests {
				doTest(&test)
			}
		})
	})
}

func TestExpenditureControllerUpdate(t *testing.T) {
	withDb(func() {
		now := time.Now()

		db.DB.Create(&models.Expenditure{
			Amount: 123,
			Date:   now,
		})

		db.DB.Create(&models.Expenditure{
			Amount: 321,
			Date:   now,
			Category: &models.Category{
				Name: "cat1",
			},
		})

		db.DB.Create(&models.Expenditure{
			Amount: 123,
			Date:   now,
			Category: &models.Category{
				Name: "cat2",
			},
		})

		db.DB.Create(&models.Expenditure{
			Amount:     123,
			Date:       now,
			CategoryID: 1,
		})

		tests := []test{}
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `{"date": "` + now.Add(24*time.Hour).Format(time.RFC3339) + `"}`,
			Params:             []string{"id"},
			ParamValues:        []string{"1"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     1,
				Amount: 123,
				Date:   now.Add(24 * time.Hour),
			},
		})
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "get",
			Endpoint:           ExpenditureController.Show,
			ContentType:        "application/json",
			Params:             []string{"id"},
			ParamValues:        []string{"1"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     1,
				Amount: 123,
				Date:   now.Add(24 * time.Hour),
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `{"amount": 321.98, "category": "cat3"}`,
			Params:             []string{"id"},
			ParamValues:        []string{"1"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     1,
				Amount: 321.98,
				Date:   now.Add(24 * time.Hour),
				Category: &CategoryResponse{
					ID:   3,
					Name: "cat3",
				},
			},
		})
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "get",
			Endpoint:           ExpenditureController.Show,
			ContentType:        "application/json",
			Params:             []string{"id"},
			ParamValues:        []string{"1"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     1,
				Amount: 321.98,
				Date:   now.Add(24 * time.Hour),
				Category: &CategoryResponse{
					ID:   3,
					Name: "cat3",
				},
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `{"category": "cat1"}`,
			Params:             []string{"id"},
			ParamValues:        []string{"2"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     2,
				Amount: 321,
				Date:   now,
				Category: &CategoryResponse{
					ID:   1,
					Name: "cat1",
				},
			},
		})
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "get",
			Endpoint:           ExpenditureController.Show,
			ContentType:        "application/json",
			Params:             []string{"id"},
			ParamValues:        []string{"2"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     2,
				Amount: 321,
				Date:   now,
				Category: &CategoryResponse{
					ID:   1,
					Name: "cat1",
				},
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `{"amount": -987.5}`,
			Params:             []string{"id"},
			ParamValues:        []string{"3"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     3,
				Amount: -987.5,
				Date:   now,
				Category: &CategoryResponse{
					ID:   2,
					Name: "cat2",
				},
			},
		})
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "get",
			Endpoint:           ExpenditureController.Show,
			ContentType:        "application/json",
			Params:             []string{"id"},
			ParamValues:        []string{"3"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     3,
				Amount: -987.5,
				Date:   now,
				Category: &CategoryResponse{
					ID:   2,
					Name: "cat2",
				},
			},
		})

		// Removing category.
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `{"category": ""}`,
			Params:             []string{"id"},
			ParamValues:        []string{"3"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     3,
				Amount: -987.5,
				Date:   now,
			},
		})
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "get",
			Endpoint:           ExpenditureController.Show,
			ContentType:        "application/json",
			Params:             []string{"id"},
			ParamValues:        []string{"3"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedExpenditureResponse: &ExpenditureResponse{
				ID:     3,
				Amount: -987.5,
				Date:   now,
			},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `{"category": "amount": 31212}`,
			Params:             []string{"id"},
			ParamValues:        []string{"3313"},
			ExpectedStatusCode: http.StatusNotFound,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `}{asa}`,
			Params:             []string{"id"},
			ParamValues:        []string{"2"},
			ExpectedStatusCode: http.StatusBadRequest,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "post",
			Endpoint:           ExpenditureController.Update,
			ContentType:        "application/json",
			PostData:           `{"amount": "abc"}`,
			Params:             []string{"id"},
			ParamValues:        []string{"3"},
			ExpectedStatusCode: http.StatusBadRequest,
		})

		Convey("Updating expenditures.", t, func() {
			for _, test := range tests {
				doTest(&test)
			}
		})
	})
}

func TestExpenditureControllerDelete(t *testing.T) {
	withDb(func() {
		Convey("Inserting expenditures.", t, func() {
			var expenditure *models.Expenditure

			expenditure = &models.Expenditure{
				Amount: 123,
				Date:   time.Now(),
			}
			db.DB.Create(expenditure)

			expenditure = &models.Expenditure{
				Amount: -1,
				Date:   time.Now().Add(24 * time.Hour),
			}
			db.DB.Create(expenditure)

			expenditure = &models.Expenditure{
				Amount:   -100.53,
				Date:     time.Now().Add(48 * time.Hour),
				Category: &models.Category{Name: "cat2"},
			}
			db.DB.Create(expenditure)

			expenditure = &models.Expenditure{
				Amount:   100.53,
				Date:     time.Now().Add(100 * time.Hour),
				Category: expenditure.Category, // Use same category as previous.
			}
			db.DB.Create(expenditure)
		})

		tests := []test{}
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "delete",
			Endpoint:           ExpenditureController.Delete,
			Params:             []string{"id"},
			ParamValues:        []string{"1"},
			ExpectedStatusCode: http.StatusOK,
		})
		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "get",
			Endpoint:           ExpenditureController.Show,
			Params:             []string{"id"},
			ParamValues:        []string{"1"},
			ExpectedStatusCode: http.StatusNotFound,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "delete",
			Endpoint:           ExpenditureController.Delete,
			Params:             []string{"id"},
			ParamValues:        []string{"1987"},
			ExpectedStatusCode: http.StatusNotFound,
		})

		tests = append(tests, test{
			URL:                          "/api/expenditures/:id",
			Method:                       "delete",
			Endpoint:                     ExpenditureController.Delete,
			Params:                       []string{"id"},
			ParamValues:                  []string{"abc"},
			ExpectedStatusCodeComparison: ShouldBeIn,
			ExpectedStatusCode:           []interface{}{http.StatusBadRequest, http.StatusNotFound},
		})

		tests = append(tests, test{
			URL:                          "/api/expenditures/:id",
			Method:                       "delete",
			Endpoint:                     ExpenditureController.Delete,
			Params:                       []string{"id"},
			ParamValues:                  []string{"0"},
			ExpectedStatusCodeComparison: ShouldBeIn,
			ExpectedStatusCode:           []interface{}{http.StatusBadRequest, http.StatusNotFound},
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "delete",
			Endpoint:           ExpenditureController.Delete,
			Params:             []string{"id"},
			ParamValues:        []string{"2"},
			ExpectedStatusCode: http.StatusOK,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "delete",
			Endpoint:           ExpenditureController.Delete,
			Params:             []string{"id"},
			ParamValues:        []string{"3"},
			ExpectedStatusCode: http.StatusOK,
		})

		tests = append(tests, test{
			URL:                "/api/expenditures/:id",
			Method:             "delete",
			Endpoint:           ExpenditureController.Delete,
			Params:             []string{"id"},
			ParamValues:        []string{"4"},
			ExpectedStatusCode: http.StatusOK,
		})

		Convey("Deleting expenditures.", t, func() {
			for _, test := range tests {
				doTest(&test)
			}
		})
	})
}
