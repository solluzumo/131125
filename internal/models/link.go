package models

type LinkList struct {
	ID        string
	LinksData []*Link
}

type Link struct {
	URL    string
	Status string
}

type LinkJson struct {
	ID        string            `json:"link_num"`
	LinksData map[string]string `json:"links"`
}
