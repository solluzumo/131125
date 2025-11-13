package dto

type LinkListResponse struct {
	LinksID string            `json:"links_num"`
	Links   map[string]string `json:"links"`
}
