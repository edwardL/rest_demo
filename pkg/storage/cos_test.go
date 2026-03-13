package storage

import (
	"context"
	"fmt"
	"testing"
)

func TestNewCos(t *testing.T) {
	s3, err := NewCos(CosConfig{
		Bucket:          "nf-1302123357",
		AccessKeySecret: "",
		AccessKeyId:     "",
		Endpoint:        "https://nf-1302123357.cos.ap-shanghai.myqcloud.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	opt := &ListObjectOpts{
		Directory: "",
		MaxKeys:   200,
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
