package proxy

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"io"
	"io/ioutil"
	"net/url"
	"os"
)

var (
	sess *session.Session
)

func init() {
	sess = session.New()
}

type s3cli struct {
	*s3.S3
}

func news3cli(cfgs ...*aws.Config) s3cli {
	return s3cli{s3.New(sess, cfgs...)}
}

func (s s3cli) get(uri *url.URL) (filepath string, err error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(uri.Host),
		Key:    aws.String(uri.Path),
	}
	resp, e := s.GetObject(input)
	if e != nil {
		filepath, err = "", e
		return
	}
	e = os.Mkdir(".cert", 0700)
	if e != nil {
		if prr, ok := e.(*os.PathError); ok {
			switch {
			case os.IsExist(prr.Err):
				break // ok, continue
			default:
				filepath, err = "", prr.Err
				return
			}
		} else {
			filepath, err = "", e
			return
		}
	}
	f, e := ioutil.TempFile(".cert", "")
	if e != nil {
		filepath, err = "", e
		return
	}
	defer f.Close()
	_, e = io.Copy(f, resp.Body)
	if e != nil {
		filepath, err = "", e
		return
	}
	filepath, err = f.Name(), nil
	return
}
