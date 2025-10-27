package space

type CreateSpaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateSpaceResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
