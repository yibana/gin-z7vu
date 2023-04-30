package utils

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// CompressText 将输入的文本字符串压缩成字节数组
func CompressText(text string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(text)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UncompressText 将输入的压缩字节数组解压缩成字符串
func UncompressText(compressed []byte) (string, error) {
	buf := bytes.NewBuffer(compressed)
	gr, err := gzip.NewReader(buf)
	if err != nil {
		return "", err
	}
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CompressBytes 将输入的字节数组压缩成字节数组
func CompressBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UncompressBytes 将输入的压缩字节数组解压缩成字节数组
func UncompressBytes(compressed []byte) ([]byte, error) {
	buf := bytes.NewBuffer(compressed)
	gr, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return ioutil.ReadAll(gr)
}
