package core

import (
	"reflect"
	"testing"
)

// X's chunk map lists locale data files that always 404 from the JS CDN.
// Fetching them slows discovery down and floods verbose logs.
func TestExtractBundleURLsSkipsLocaleAndEmojiBundles(t *testing.T) {
	html := `<html><head>
<script src="https://abs.twimg.com/responsive-web/client-web/emoji.fr.js"></script>
<script src="https://abs.twimg.com/responsive-web/client-web/en-GB.js"></script>
<script src="https://abs.twimg.com/responsive-web/client-web/fr.js"></script>
<script src="https://abs.twimg.com/responsive-web/client-web/ar-x-fm.js"></script>
<script src="https://abs.twimg.com/responsive-web/client-web/index.node.js"></script>
<script src="https://abs.twimg.com/responsive-web/client-web/main.abc123.js"></script>
<script src="https://abs.twimg.com/x-web/x-web/entry-client-abc123.js"></script>
</head></html>`

	got := (&EndpointDiscovery{}).extractBundleURLs(html)
	want := []string{
		"https://abs.twimg.com/responsive-web/client-web/main.abc123.js",
		"https://abs.twimg.com/x-web/x-web/entry-client-abc123.js",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractBundleURLs() = %#v, want %#v", got, want)
	}
}

func TestExtractReferencedBundleURLsSkipsLocaleAssets(t *testing.T) {
	js := `const deps=["assets/emoji.en.js","assets/pt-BR.js","assets/ar-x-fm.js","./assets/chunk-abc123.js"];`

	got := extractReferencedBundleURLs(js, "https://abs.twimg.com/x-web/x-web/entry.js")
	want := []string{"https://abs.twimg.com/x-web/x-web/assets/chunk-abc123.js"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractReferencedBundleURLs() = %#v, want %#v", got, want)
	}
}
