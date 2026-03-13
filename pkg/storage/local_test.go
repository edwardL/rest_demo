package storage

import (
	"context"
	"fmt"
	"testing"
)

func TestNewLocal(t *testing.T) {
	s3, err := NewLocal(LocalConfig{
		Root:     "/Users/wuxin/worker/hobi/common",
		Endpoint: "https://upload.chatie.live",
	})
	if err != nil {
		t.Fatal(err)
	}
	opt := &ListObjectOpts{
		Directory: "",
		MaxKeys:   5,
	}
	data := []string{}
	for {
		res, err := s3.ListObjects(context.Background(), opt)
		if err != nil {
			t.Fatal(err)
			return
		}
		if !res.HasMore {
			break
		}
		for _, item := range res.List {
			data = append(data, item.Url)
		}
		opt.NextToken = res.NextToken
		break
	}
	for i, item := range data {
		fmt.Println(i, item)
	}
}
