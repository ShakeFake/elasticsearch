package model

type Caption struct {
	Text string `json:"text"`
}

type Student struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type OnceAggregation struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}

type SearchRecordItem struct {
	Flag    []string `json:"flag"`
	Command string   `json:"command"`
}

// EntryItem 是否需要添加进度什么的。或者通过 notion 文章映射得了。
type EntryItem struct {
	Flag []string `json:"flag"`
	Link string   `json:"link"`
}
