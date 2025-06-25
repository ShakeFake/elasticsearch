package utils

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
)

// highlight 链接
// https://www.elastic.co/guide/en/elasticsearch/reference/current/highlighting.html

func SpecialField(index string) {
	query := elastic.NewBoolQuery()

	hightlight := elastic.NewHighlight().
		PreTags("<b class='highlight'>").
		PostTags("</b>").
		Fields(
			elastic.NewHighlighterField("name"),
			elastic.NewHighlighterField("age"),
		)

	res, err := ESCLI.Search().Index(index).Query(query).Highlight(hightlight).Do(context.Background())
	if err != nil {

	}

	for _, hit := range res.Hits.Hits {
		fmt.Println(string(hit.Source))

		// 按照 field 的总结的 highlight
		fmt.Println(hit.Highlight)
	}

}
