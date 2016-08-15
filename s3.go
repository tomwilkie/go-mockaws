package mockaws

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
)

type mockS3 struct {
	mtx     sync.RWMutex
	buckets map[string]*mockS3Bucket
}

type mockS3Bucket struct {
	objects map[string][]byte
}

func newMockS3() *mockS3 {
	return &mockS3{
		buckets: map[string]*mockS3Bucket{},
	}
}

func (m *mockS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	buf, err := ioutil.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	bucket, ok := m.buckets[*input.Bucket]
	if !ok {
		bucket = &mockS3Bucket{
			objects: map[string][]byte{},
		}
		m.buckets[*input.Bucket] = bucket
	}

	bucket.objects[*input.Key] = buf
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	bucket, ok := m.buckets[*input.Bucket]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	buf, ok := bucket.objects[*input.Key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewBuffer(buf)),
	}, nil
}
