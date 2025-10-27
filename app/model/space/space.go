package space

type CreateSpaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Owner      string `json:"owner" binding:"required"`
	Permission  int    `json:"permission" binding:"required"`
	Description string `json:"description"`
}

type CreateSpaceResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Owner      string `json:"owner"`
	Permission  int    `json:"permission"`
	Description string `json:"description"`
}


type ListSpaceFilter struct {
	Name       string
	Permission int
	Owner      string
}

type SpaceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Owner      string `json:"owner"`
	Permission  int    `json:"permission"`
	Description string `json:"description"`
}
type UpdateSpaceRequest struct {
	Name        *string `json:"name"`
	Owner      *string `json:"owner"`
	Permission  *int    `json:"permission"`
	Description *string `json:"description"`
}

