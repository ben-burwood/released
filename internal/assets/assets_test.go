package assets

import "testing"

func TestRewrite(t *testing.T) {
	body := `![x](https://github.com/user-attachments/assets/abc-123) <img src="https://github.com/user-attachments/assets/DEF"> https://github.com/other/assets/abc`
	want := `![x](/api/assets/abc-123) <img src="/api/assets/DEF"> https://github.com/other/assets/abc`
	if got := Rewrite(body); got != want {
		t.Errorf("Rewrite = %q, want %q", got, want)
	}
}
