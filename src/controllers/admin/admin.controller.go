package admin

type (
	apiResponse map[string]interface{}
)

var (
	internalServerError = apiResponse{
		"status":  "failed",
		"message": "error occcured",
	}
)
