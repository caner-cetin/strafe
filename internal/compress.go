package internal

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"io"
)

func CompressJSON(data []byte) ([]byte, error) {
	base64String := base64.StdEncoding.EncodeToString(data)
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(base64String); err != nil {
		return nil, err
	}
	jsonString := bytes.TrimSpace(buffer.Bytes())
	var compressed bytes.Buffer
	w, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(jsonString); err != nil {
		w.Close()
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return compressed.Bytes(), nil
}

func DecompressJSON[T any](compressed []byte, target *T) error {
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	decompressed = bytes.TrimSpace(decompressed)
	var base64String string
	decoder := json.NewDecoder(bytes.NewReader(decompressed))
	decoder.UseNumber()
	if err := decoder.Decode(&base64String); err != nil {
		return err
	}
	jsonData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return err
	}
	decoder = json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	return decoder.Decode(target)
}
