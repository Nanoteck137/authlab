package types

const (
	RoleSuperUser = "super_user"
	RoleAdmin     = "admin"
)

type Page struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type Images struct {
	Original string `json:"original"`
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
}
