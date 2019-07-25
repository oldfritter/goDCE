package utils

import (
	"fmt"
)

type Response struct {
	Head map[string]string `json:"head"`
	Body interface{}       `json:"body"`
}

type ArrayBodyStruct struct {
	CurrentPage  int         `json:"current_page"`
	TotalPages   int         `json:"total_pages"`
	PerPage      int         `json:"per_page"`
	NextPage     int         `json:"next_page"`
	PreviousPage int         `json:"previous_page"`
	Data         interface{} `json:"data"`
}

type ArrayDataResponse struct {
	Head map[string]string `json:"head"`
	Body interface{}       `json:"body"`
}

var (
	SuccessResponse = Response{Head: map[string]string{"code": "1000", "msg": "Success."}}
	ArrayResponse   = ArrayDataResponse{Head: map[string]string{"code": "1000", "msg": "Success."}}
)

func BuildError(code string) Response {
	return Response{Head: map[string]string{"code": code}}
}

func (errorResponse Response) Error() string {
	return fmt.Sprintf("code: %s; msg: %s", errorResponse.Head["code"], errorResponse.Head["msg"])
}

func (arrayResponse *ArrayDataResponse) Init(data interface{}, page, count, per_page int) {
	total_page := count / per_page
	if (count % per_page) != 0 {
		total_page += 1
	}

	nextPage := page + 1
	if nextPage > total_page {
		nextPage = total_page
	}
	previousPage := page - 1
	if previousPage < 1 {
		previousPage = 1
	}

	body := ArrayBodyStruct{}

	body.Data = data
	body.CurrentPage = page
	body.TotalPages = total_page
	body.PerPage = per_page
	body.NextPage = nextPage
	body.PreviousPage = previousPage

	arrayResponse.Body = body
}
