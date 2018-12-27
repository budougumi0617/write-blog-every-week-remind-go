package rss

import (
	time "time"

	database "../database"
	"../date"
	gofeed "github.com/mmcdole/gofeed"
)

// FindTargetUserList ブログを書いていないユーザーを取得する
func FindTargetUserList(allMemberDataList []database.WriteBlogEveryWeek, targetMonday time.Time) map[string]int {
	// 日本時間に合わせる
	locale, _ := time.LoadLocation("Asia/Tokyo")
	parser := gofeed.NewParser()

	results := make(map[string]int)
	for _, wbem := range allMemberDataList {
		// フィードを取得
		feed, err := parser.ParseURL(wbem.FeedURL)
		if err != nil {
			panic("フィードが取得できませんでした。失敗したフィードURL => " + wbem.FeedURL)
		}

		if _, ok := results[wbem.UserID]; !ok {
			// データがない場合は初期化
			results[wbem.UserID] = 0
		}

		for i := 0; i < wbem.RequireCount; i++ {
			// 最新フィードの公開日を取得する
			latestPublishDate := getLatestFeedPubDate(feed, i, parser, locale)

			// 今週の月曜日が過去ではない場合は、まだ今週ブログを書いていない
			if !targetMonday.Before(latestPublishDate) {
				results[wbem.UserID]++
			}
		}
	}

	return results
}

// getLatestFeedPubDate 最新フィードの公開日を取得する
func getLatestFeedPubDate(feed *gofeed.Feed, requireCount int, parser *gofeed.Parser, locale *time.Location) time.Time {
	if (requireCount + 1) > len(feed.Items) {
		// そもそも記事数が足りない場合は公開日を取得できないのでlatestは、必ず通知対象となる今週の月曜日と合わせる
		return date.GetThisMonday()
	}

	// 最新日を取得
	published := feed.Items[requireCount].Published
	latest, err := time.ParseInLocation(time.RFC3339, published, locale)
	if err != nil {
		// 取得できない = フォーマットを変えれば取得できる可能性がある
		latest2, err := time.ParseInLocation(time.RFC1123Z, published, locale)
		if err != nil {
			// それでも取得できない場合は、フィードで取得した生データをもらう
			latest = *feed.Items[requireCount].PublishedParsed
		} else {
			latest = latest2
		}
	}

	return latest
}
