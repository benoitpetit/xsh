package core

import "testing"

func TestParseTweetVolume(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{name: "plain number", text: "1,234 posts", want: 1234},
		{name: "k suffix", text: "12.5K posts", want: 12500},
		{name: "m suffix", text: "1.2M tweets", want: 1200000},
		{name: "invalid", text: "no volume", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTweetVolume(tt.text)
			if got != tt.want {
				t.Fatalf("parseTweetVolume(%q) = %d, want %d", tt.text, got, tt.want)
			}
		})
	}
}

func TestParseTrendsFromExploreResponse(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"explore_page": map[string]interface{}{
				"body": map[string]interface{}{
					"initialTimeline": map[string]interface{}{
						"timeline": map[string]interface{}{
							"timeline": map[string]interface{}{
								"instructions": []interface{}{
									map[string]interface{}{
										"entries": []interface{}{
											map[string]interface{}{
												"content": map[string]interface{}{
													"items": []interface{}{
														map[string]interface{}{
															"item": map[string]interface{}{
																"itemContent": map[string]interface{}{
																	"__typename": "TimelineTrend",
																	"name":       "#GoLang",
																	"social_context": map[string]interface{}{
																		"text": "Trending in Tech · 12.5K posts",
																	},
																	"trend_metadata": map[string]interface{}{
																		"url": map[string]interface{}{"url": "/search?q=%23GoLang"},
																	},
																	"promoted_content": map[string]interface{}{},
																},
															},
														},
														map[string]interface{}{
															"item": map[string]interface{}{
																"itemContent": map[string]interface{}{
																	"trend": map[string]interface{}{
																		"name": "#OpenSource",
																		"trendMetadata": map[string]interface{}{
																			"metaDescription": "8,765 posts",
																		},
																		"url": map[string]interface{}{"url": "/search?q=%23OpenSource"},
																	},
																	"trendContext": map[string]interface{}{
																		"text": "Technology",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	trends := parseTrendsFromExploreResponse(data)
	if len(trends) != 2 {
		t.Fatalf("expected 2 trends, got %d", len(trends))
	}

	if trends[0].Name != "#GoLang" {
		t.Fatalf("first trend name = %q, want %q", trends[0].Name, "#GoLang")
	}
	if trends[0].TweetVolume != 12500 {
		t.Fatalf("first trend volume = %d, want %d", trends[0].TweetVolume, 12500)
	}
	if !trends[0].IsPromoted {
		t.Fatalf("first trend should be marked promoted")
	}

	if trends[1].Name != "#OpenSource" {
		t.Fatalf("second trend name = %q, want %q", trends[1].Name, "#OpenSource")
	}
	if trends[1].TweetVolume != 8765 {
		t.Fatalf("second trend volume = %d, want %d", trends[1].TweetVolume, 8765)
	}
	if trends[1].Rank != 2 {
		t.Fatalf("second trend rank = %d, want %d", trends[1].Rank, 2)
	}
}
