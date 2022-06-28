package internal

type MenuItem struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Tooltip string `json:"tooltip"`
}

type MenuItemReply struct {
	MenuItem
	Id int `json:"id"`
}

type MenuItemClickEvent struct {
	MenuItem
	Id int `json:"id"`
}
