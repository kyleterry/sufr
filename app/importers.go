package app

import (
	"encoding/json"
	"strings"
	"sync"
)

func shitbucketImporter(b []byte) (int, error) {
	var payload struct {
		Urls  []map[string]interface{}
		Count int
	}

	if err := json.Unmarshal(b, &payload); err != nil {
		return 0, err
	}

	var wg sync.WaitGroup

	wg.Add(payload.Count)
	for _, item := range payload.Urls {
		go func(item map[string]interface{}) {
			defer wg.Done()
			url := URL{}
			url.URL = item["url"].(string)
			url.Title = item["url_title"].(string)
			var tags []string
			itags := item["tags"]
			if itags != nil {
				for _, t := range itags.([]interface{}) {
					tags = append(tags, t.(string))
				}
			}
			url.SaveWithTags(strings.Join(tags, " "))
		}(item)
	}

	wg.Wait()

	return payload.Count, nil
}
