package common

type MessageType byte

const (
	AddItemMessage MessageType = iota + 1
	RemoveItemMessage
	GetItemMessage
	GetAllItemsMessage
)

type Message struct {
	//Sender string      `json:"sender"`
	Type MessageType `json:"type"`
	Data []byte      `json:"data"`
}

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AddItem struct {
	TheItem Item `json:"item"`
}

type RemoveItem struct {
	Key string `json:"key"`
}

type GetItem struct {
	Key string `json:"key"`
}

type GetAllItems struct {
}
