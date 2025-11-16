package models

type LinkList struct {
	ID        string
	LinksData map[string]string
}

func NewLinkList(id string, data map[string]string) *LinkList {
	return &LinkList{
		ID:        id,
		LinksData: data,
	}
}

type LinkJson struct {
	ID        string            `json:"link_num"`
	LinksData map[string]string `json:"links"`
}
