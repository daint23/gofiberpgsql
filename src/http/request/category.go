package request

type CategoryCreateRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description"`
}

type CategoryUpdateRequest struct {
	Id          int    `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description"`
}

type CategoryQueryParams struct {
	Id    int
	Limit int
}
