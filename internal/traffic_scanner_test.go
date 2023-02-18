package internal

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestScanTraffics(t *testing.T) {
	files, err := ListTrafficsFiles("../", "traffics.log")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(files)

	for _, file := range files {
		fmt.Println(file)
		testScanTraffics(t, file)
	}
}

func testScanTraffics(t *testing.T, filePath string) {
	scanner, err := NewTrafficsScanner(filePath)
	if err != nil {
		t.Fatal(err)
	}

	err = scanner.Scan(context.Background(), func(time time.Time, identifier string, bytes int64) {
		fmt.Println(time, identifier, bytes)
	})
	if err != nil {
		t.Fatal(err)
	}
}
