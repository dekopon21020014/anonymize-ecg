package mfer

import (
	"bytes"
	"testing"
)

func TestAnonymize(t *testing.T) {
	// テストデータ
	testData := []byte{
		// プリアンブル
		0x40, 0x20, 0x4d, 0x46, 0x52, 0x20, 0x53, 0x74, 0x61, 0x6E, 0x64, 0x61, 0x72, 0x64, 0x20, 0x31, 0x32, 0x20, 0x6C, 0x65, 0x61, 0x64, 0x73, 0x20, 0x45, 0x43, 0x47, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
		// バイトオーダー
		0x01, 0x01, 0x01,
		// 文字コード
		0x03, 0x08, 0x55, 0x4E, 0x49, 0x43, 0x4F, 0x44, 0x45, 0x00,
		// 患者名
		0x81, 0x2a, 0xe5, 0x85, 0x89, 0xe9, 0x9b, 0xbb, 0xe3, 0x80, 0x80, 0xe8, 0x8a, 0xb1, 0xe5, 0xad, 0x90, 0x5e, 0xef, 0xbd, 0xba, 0xef, 0xbd, 0xb3, 0xef, 0xbe, 0x83, 0xef, 0xbe, 0x9e, 0xef, 0xbe, 0x9d, 0x20, 0xef, 0xbe, 0x8a, 0xef, 0xbe, 0x85, 0xef, 0xbd, 0xba, 0x00,
		// 患者ID
		0x82, 0x0b, 0x31, 0x31, 0x32, 0x33, 0x37, 0x30, 0x30, 0x30, 0x35, 0x31, 0x00,
		// 年齢・生年月日
		0x83, 0x07, 0x16, 0xfe, 0x1f, 0xc0, 0x07, 0x0b, 0x17,
	}

	/*
	 * 関数の返り値として期待されるデータ
	 * - 患者名が削除される
	 * - 患者IDが削除される
	 * - 日齢が0になる
	 * - 生年月日の日が0になる(生年月はそのまま)
	 */
	expectedData := []byte{
		// プリアンブル
		0x40, 0x20, 0x4d, 0x46, 0x52, 0x20, 0x53, 0x74, 0x61, 0x6E, 0x64, 0x61, 0x72, 0x64, 0x20, 0x31, 0x32, 0x20, 0x6C, 0x65, 0x61, 0x64, 0x73, 0x20, 0x45, 0x43, 0x47, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
		// バイトオーダー
		0x01, 0x01, 0x01,
		// 文字コード
		0x03, 0x08, 0x55, 0x4E, 0x49, 0x43, 0x4F, 0x44, 0x45, 0x00,
		// 患者名
		// 0x81, 0x2a, 0xe5, 0x85, 0x89, 0xe9, 0x9b, 0xbb, 0xe3, 0x80, 0x80, 0xe8, 0x8a, 0xb1, 0xe5, 0xad, 0x90, 0x5e, 0xef, 0xbd, 0xba, 0xef, 0xbd, 0xb3, 0xef, 0xbe, 0x83, 0xef, 0xbe, 0x9e, 0xef, 0xbe, 0x9d, 0x20, 0xef, 0xbe, 0x8a, 0xef, 0xbe, 0x85, 0xef, 0xbd, 0xba, 0x00,
		// 患者ID
		// 0x82, 0x0b, 0x31, 0x31, 0x32, 0x33, 0x37, 0x30, 0x30, 0x30, 0x35, 0x31, 0x00,
		// 年齢・生年月日
		// 0x83, 0x07, 0x16, 0x00, 0x00, 0xc0, 0x07, 0x0b, 0x00,
	}

	got, err := Anonymize(testData)
	if err != nil {
		t.Fatal(err)
	}

	// テストの判定
	if !bytes.Equal(expectedData, got) {
		t.Fatalf("expected: %v, got %v", expectedData, got)
	}
}