package database

type okxWalletBalance struct {
	Id          int64  `json:"id"`
	Address     string `json:"address"`
	AddressName string `json:"address_name"`

	Balance     string `json:"balance"`
	RecordDate  string `json:"record_date"`
	Direction   string `json:"direction"`
	ChangeRange string `json:"change_range"`
	CreateTime  int64  `json:"create_time"`
	UpdateTime  int64  `json:"update_time"`
}
