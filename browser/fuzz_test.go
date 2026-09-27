package browser

import "testing"

func FuzzCookieAccumulator(f *testing.F) {
	for _, seed := range []struct {
		host, name, value string
	}{
		{"x.com", "auth_token", "token"},
		{".twitter.com", "ct0", "csrf"},
		{"", "", ""},
	} {
		f.Add(seed.host, seed.name, seed.value)
	}
	f.Fuzz(func(t *testing.T, host, name, value string) {
		acc := newCookieAccumulator()
		acc.add(host, name, value, nil)
		_, _ = acc.credentials("fuzz")
		_ = preferCookieHost(host, host, value, value)
	})
}

func FuzzChromeCBCInput(f *testing.F) {
	f.Add([]byte("key"), []byte("iv"), []byte("ciphertext"))
	f.Fuzz(func(t *testing.T, key, iv, ciphertext []byte) {
		_, _ = aesCBCDecrypt(key, iv, ciphertext)
	})
}
