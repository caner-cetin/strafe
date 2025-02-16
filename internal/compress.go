package internal

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
)

func CompressJSON(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var compressed bytes.Buffer
	w, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(jsonData); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return compressed.Bytes(), nil
}

func DecompressJSON(compressed []byte, target interface{}) error {
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	defer r.Close()
	decompressed, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(decompressed, target)
}
